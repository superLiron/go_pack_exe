package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Read body failed", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req RequestBody
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Webhook == "" || req.Msg == "" {
		http.Error(w, "缺少地址或者内容", http.StatusBadRequest)
		return
	}


	if !strings.HasPrefix(req.Webhook, "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=") {
		http.Error(w, "Invalid webhook URL", http.StatusBadRequest)
		fmt.Println("非法地址")
		return
	}

	// ✅ 关键修改：自己构造企业微信需要的 JSON
	buildMsg := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": req.Msg,
		},
	}

	weChatMsg, err := json.Marshal(buildMsg)
	if err != nil {
		http.Error(w, "Build message failed", http.StatusInternalServerError)
		fmt.Printf("❌ 消息构造失败: %v\n", err)
		return
	}
	

	resp, err := http.Post(req.Webhook, "application/json", strings.NewReader(string(weChatMsg)))
	// resp, err := http.Post(req.Webhook, "application/json", strings.NewReader(string(body)))
	if err != nil {
		http.Error(w, "Forward failed: "+err.Error(), http.StatusInternalServerError)
		fmt.Printf("❌ 转发失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

func main() {
	http.HandleFunc("/send", sendHandler)
	fmt.Println("🚀 企业微信代理服务启动成功！")
	fmt.Println("监听页面: http://localhost:8081")
	fmt.Println("请保持此窗口打开...")
	err := http.ListenAndServe("127.0.0.1:8081", nil)
	if err != nil {
		fmt.Printf("❌ 启动失败: %v\n", err)
	}
}
