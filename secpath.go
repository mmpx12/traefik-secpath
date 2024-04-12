package traefik_secpath

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Config struct {
	Rules map[string]RuleConfig
}

type RuleConfig struct {
	PathRule string `json:"pathRule"`
	IPRule   string `json:"ipRule"`
	TypeRule string `json:"typeRule"`
	NewPath  string `json:"newPath"`
	PathReg  *regexp.Regexp
}

func Print(str string, a ...interface{}) {
	fmt.Fprintf(os.Stdout, str, a...)
	os.Stdout.Write([]byte("\n"))
}

func CreateConfig() *Config {
	return &Config{
		Rules: make(map[string]RuleConfig),
	}
}

type Moni struct {
	next  http.Handler
	name  string
	rules map[string]RuleConfig
}

func checkCIDR(cidr string, target string) bool {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	targetIP := net.ParseIP(target)
	if targetIP == nil {
		return false
	}
	return ipNet.Contains(targetIP)
}

func checkRange(ipRange string, target string) bool {
	ipRangeParts := strings.Split(ipRange, "-")
	if len(ipRangeParts) != 2 {
		return false
	}
	startIP := net.ParseIP(strings.TrimSpace(ipRangeParts[0]))
	endIP := net.ParseIP(strings.TrimSpace(ipRangeParts[1]))
	targetIP := net.ParseIP(target)
	if startIP == nil || endIP == nil || targetIP == nil {
		return false
	}
	start := binary.BigEndian.Uint32(startIP.To4())
	end := binary.BigEndian.Uint32(endIP.To4())
	targetInt := binary.BigEndian.Uint32(targetIP.To4())
	return start <= targetInt && targetInt <= end
}

func check(cidrOrRange string, target string) bool {
	var ips []string

	if strings.Contains(cidrOrRange, ",") {
		ips = strings.Split(cidrOrRange, ",")
	} else {
		ips = []string{cidrOrRange}
	}

	for _, ip := range ips {
		if strings.Contains(ip, "/") {
			if checkCIDR(ip, target) {
				return true
			}
		} else if strings.Contains(ip, "-") {
			if checkRange(ip, target) {
				return true
			}
		} else {
			IP := net.ParseIP(ip)
			if IP != nil && IP.String() == target {
				return true
			}
		}
	}

	return false
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	for ruleName, rule := range config.Rules {
		Print("Rule:\n  %s:\n    IP: %s\n    Path: %s\n    Type: %s", ruleName, rule.IPRule, rule.PathRule, rule.TypeRule)
		reg, err := regexp.Compile(rule.PathRule) //rule.PathRule)
		if err != nil {
			Print("pathreg panic")
			panic(err)
		}
		rule.PathReg = reg
		config.Rules[ruleName] = rule
		if rule.NewPath != "" {
			Print("    New Path: %s", rule.NewPath)
		}
	}

	Print("")

	return &Moni{
		next:  next,
		rules: config.Rules,
	}, nil
}

func (a *Moni) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for ruleName, rule := range a.rules {
		if rule.PathReg.MatchString(req.URL.Path) {
			if (rule.TypeRule == "block" && check(rule.IPRule, req.Header.Get(`X-Real-Ip`))) ||
				(rule.TypeRule == "allow" && !check(rule.IPRule, req.Header.Get(`X-Real-Ip`))) ||
				(rule.TypeRule == "" && !check(rule.IPRule, req.Header.Get(`X-Real-Ip`))) {
				Print("Matched rule %s for path %s and IP %s [%s] %s", ruleName, rule.PathRule, req.Header.Get(`X-Real-Ip`), rule.IPRule, rule.TypeRule)
				rw.WriteHeader(http.StatusUnauthorized)
				rw.Write([]byte("Not welcome\n"))
				return
			} else if rule.TypeRule == "fake" && check(rule.IPRule, req.Header.Get(`X-Real-Ip`)) {
				Print("Matched rule %s for path %s and IP %s [%s] %s", ruleName, rule.PathRule, req.Header.Get(`X-Real-Ip`), rule.IPRule, rule.TypeRule)
				Print("Change path %s to %s", req.URL.Path, rule.NewPath)
				req.RequestURI = rule.PathReg.ReplaceAllString(req.RequestURI, rule.NewPath) //rule.NewPath
				break
			} else if rule.TypeRule == "redirection" && check(rule.IPRule, req.Header.Get(`X-Real-Ip`)) {
				Print("Matched rule %s for path %s and IP %s [%s] %s", ruleName, rule.PathRule, req.Header.Get(`X-Real-Ip`), rule.IPRule, rule.TypeRule)
				http.Redirect(rw, req, rule.NewPath, http.StatusFound)
				return
			}
			Print("Matched rule %s for path %s and IP %s [%s] %s", ruleName, rule.PathRule, req.Header.Get(`X-Real-Ip`), rule.IPRule, rule.TypeRule)
			break
		}
	}

	a.next.ServeHTTP(rw, req)
}
