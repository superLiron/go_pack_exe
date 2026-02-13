package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type RequestBody struct {
	Webhook string `json:"webhook"`
	Msg     string `json:"msg"`
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" && r.Method != "OPTIONS" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 记录请求时间
	start := time.Now()
	fmt.Printf("📥 接收请求: %s\n", r.URL.Path)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Read body failed", http.StatusBadRequest)
		fmt.Printf("❌ 读取请求体失败: %v\n", err)
		return
	}
	defer r.Body.Close()

	var req RequestBody
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		fmt.Printf("❌ JSON 解析失败: %v\n", err)
		return
	}

	fmt.Printf("📤 准备转发: webhook=%s, msg=%s\n", req.Webhook, req.Msg)

	if req.Webhook == "" || req.Msg == "" {
		http.Error(w, "Missing webhook or msg", http.StatusBadRequest)
		fmt.Println("❌ 缺少 webhook 或 msg")
		return
	}

	if !strings.HasPrefix(req.Webhook, "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=") {
		http.Error(w, "Invalid webhook URL", http.StatusBadRequest)
		fmt.Println("❌ 非法 webhook URL")
		return
	}

	// 发送请求到企业微信
	resp, err := http.Post(req.Webhook, "application/json", strings.NewReader(string(body)))
	if err != nil {
		http.Error(w, "Forward failed: "+err.Error(), http.StatusInternalServerError)
		fmt.Printf("❌ 转发失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("✅ 转发成功! 状态码: %d, 响应: %s\n", resp.StatusCode, string(respBody))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

func main() {
	fmt.Println("🚀 企业微信代理服务启动成功！")
	fmt.Println("监听页面: http://localhost:8081")
	fmt.Println("请保持此窗口打开...")
	err := http.ListenAndServe("127.0.0.1:8081", nil)
	if err != nil {
		fmt.Printf("❌ 启动失败: %v\n", err)
	}
}
