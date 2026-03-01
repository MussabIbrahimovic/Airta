package stdlib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Node struct {
	Kind, Tag, Text string
	Props           map[string]any
	Kids            []any
}

func Inject(def func(string, any), args []string) {
	def("print", func(v ...any) (any, error) { fmt.Println(v...); return nil, nil })
	def("now", func(v ...any) (any, error) { return float64(time.Now().UnixMilli()), nil })
	def("sleep", func(v ...any) (any, error) {
		d := time.Duration(toNum(v, 0)) * time.Millisecond
		time.Sleep(d)
		return nil, nil
	})
	def("env", func(v ...any) (any, error) {
		if len(v) == 0 {
			return "", nil
		}
		n := fmt.Sprint(v[0])
		if x, ok := os.LookupEnv(n); ok {
			return x, nil
		}
		if len(v) > 1 {
			return v[1], nil
		}
		return nil, nil
	})
	def("json", map[string]any{"parse": func(v ...any) (any, error) {
		var out any
		if err := json.Unmarshal([]byte(fmt.Sprint(v[0])), &out); err != nil {
			return nil, err
		}
		return out, nil
	}, "stringify": func(v ...any) (any, error) { b, _ := json.Marshal(v[0]); return string(b), nil }})
	def("path", map[string]any{"join": func(v ...any) (any, error) {
		s := []string{}
		for _, x := range v {
			s = append(s, fmt.Sprint(x))
		}
		return filepath.Join(s...), nil
	}})
	def("fs", map[string]any{"read_text": func(v ...any) (any, error) { b, er := os.ReadFile(fmt.Sprint(v[0])); return string(b), er }, "write_text": func(v ...any) (any, error) {
		return nil, os.WriteFile(fmt.Sprint(v[0]), []byte(fmt.Sprint(v[1])), 0644)
	}, "mkdir": func(v ...any) (any, error) { return nil, os.MkdirAll(fmt.Sprint(v[0]), 0755) }})
	def("ui", map[string]any{"element": func(v ...any) (any, error) {
		tag := fmt.Sprint(v[0])
		props := map[string]any{}
		if len(v) > 1 {
			if m, ok := v[1].(map[string]any); ok {
				props = m
			}
		}
		kids := []any{}
		if len(v) > 2 {
			if a, ok := v[2].([]any); ok {
				kids = a
			}
		}
		return Node{Kind: "el", Tag: tag, Props: props, Kids: kids}, nil
	}, "text": func(v ...any) (any, error) { return Node{Kind: "txt", Text: fmt.Sprint(v[0])}, nil }, "renderToHtml": func(v ...any) (any, error) { return render(v[0]), nil }})
	def("http", map[string]any{"server": func(v ...any) (any, error) { return newServer(), nil }})
	def("runtime", map[string]any{"args": func(v ...any) (any, error) {
		a := []any{}
		for _, x := range args {
			a = append(a, x)
		}
		return a, nil
	}})
}

func newServer() map[string]any {
	type route struct {
		m, p string
		h    any
	}
	routes := []route{}
	mw := []any{}
	mux := http.NewServeMux()
	h := map[string]any{}
	h["use"] = func(v ...any) (any, error) { mw = append(mw, v[0]); return nil, nil }
	h["get"] = func(v ...any) (any, error) {
		routes = append(routes, route{"GET", fmt.Sprint(v[0]), v[1]})
		return nil, nil
	}
	h["post"] = func(v ...any) (any, error) {
		routes = append(routes, route{"POST", fmt.Sprint(v[0]), v[1]})
		return nil, nil
	}
	h["listen"] = func(v ...any) (any, error) {
		port := fmt.Sprint(v[0])
		if !strings.HasPrefix(port, ":") {
			port = ":" + port
		}
		for _, r := range routes {
			rr := r
			mux.HandleFunc(rr.p, func(w http.ResponseWriter, req *http.Request) {
				if req.Method != rr.m {
					return
				}
				reqObj := map[string]any{"method": req.Method, "path": req.URL.Path, "headers": map[string]any{}}
				resObj := map[string]any{}
				resObj["status"] = func(v ...any) (any, error) { w.WriteHeader(int(toNum(v, 0))); return resObj, nil }
				resObj["json"] = func(v ...any) (any, error) {
					w.Header().Set("Content-Type", "application/json")
					b, _ := json.Marshal(v[0])
					_, _ = w.Write(b)
					return nil, nil
				}
				resObj["html"] = func(v ...any) (any, error) {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					_, _ = w.Write([]byte(render(v[0])))
					return nil, nil
				}
				for _, m := range mw {
					if f, ok := m.(func(...any) (any, error)); ok {
						_, _ = f(reqObj, resObj, func(v ...any) (any, error) { return nil, nil })
					}
				}
				if f, ok := rr.h.(func(...any) (any, error)); ok {
					_, _ = f(reqObj, resObj)
				}
			})
		}
		return nil, http.ListenAndServe(port, mux)
	}
	return h
}

func render(v any) string {
	switch t := v.(type) {
	case Node:
		if t.Kind == "txt" {
			return htmlEsc(t.Text)
		}
		attrs := []string{}
		for k, val := range t.Props {
			if k == "style" {
				if m, ok := val.(map[string]any); ok {
					pairs := []string{}
					for sk, sv := range m {
						pairs = append(pairs, fmt.Sprintf("%s:%v", sk, sv))
					}
					attrs = append(attrs, fmt.Sprintf("style=\"%s\"", strings.Join(pairs, ";")))
					continue
				}
			}
			attrs = append(attrs, fmt.Sprintf("%s=\"%v\"", k, val))
		}
		body := ""
		for _, c := range t.Kids {
			body += render(c)
		}
		if len(attrs) > 0 {
			return fmt.Sprintf("<%s %s>%s</%s>", t.Tag, strings.Join(attrs, " "), body, t.Tag)
		}
		return fmt.Sprintf("<%s>%s</%s>", t.Tag, body, t.Tag)
	case []any:
		s := ""
		for _, x := range t {
			s += render(x)
		}
		return s
	default:
		return htmlEsc(fmt.Sprint(v))
	}
}
func htmlEsc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	return strings.ReplaceAll(s, ">", "&gt;")
}
func toNum(v []any, idx int) float64 {
	if len(v) <= idx {
		return 0
	}
	switch t := v[idx].(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case string:
		var f float64
		fmt.Sscanf(t, "%f", &f)
		return f
	default:
		return 0
	}
}
