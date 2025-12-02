#!/bin/bash

# 测试前缀路由功能

BASE_URL="http://localhost:8080"
API_KEY="sk-openbridge-key-1"

echo "=== 测试 1: 获取模型列表 ==="
echo "请求: GET /v1/models"
curl -s -X GET "$BASE_URL/v1/models" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" | jq '.data[] | {id, owned_by}' | head -20

echo ""
echo "=== 测试 2: 使用带前缀的模型名称请求 ==="
echo "请求: POST /v1/chat/completions with model: openai/gpt-4"
curl -s -X POST "$BASE_URL/v1/chat/completions" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openai/gpt-4",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ],
    "max_tokens": 10
  }' | jq '.model'

echo ""
echo "=== 测试 3: 使用另一个 Provider 的模型 ==="
echo "请求: POST /v1/chat/completions with model: claude/claude-3-opus-20240229"
curl -s -X POST "$BASE_URL/v1/chat/completions" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude/claude-3-opus-20240229",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ],
    "max_tokens": 10
  }' | jq '.model'

echo ""
echo "=== 测试完成 ==="
