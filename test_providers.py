#!/usr/bin/env python3
"""
OpenBridge Provider 测试脚本
演示如何使用不同的 Provider (OpenAI, Claude, Gemini)
"""

import os
from openai import OpenAI

# OpenBridge 配置
OPENBRIDGE_URL = "http://localhost:8080/v1"
OPENBRIDGE_KEY = "sk-openbridge-test-key-1"

def test_provider(model_name, prompt="你好！请简单介绍一下你自己。"):
    """测试指定的模型"""
    print(f"\n{'='*60}")
    print(f"测试模型: {model_name}")
    print(f"{'='*60}")
    
    client = OpenAI(
        api_key=OPENBRIDGE_KEY,
        base_url=OPENBRIDGE_URL
    )
    
    try:
        print(f"发送请求: {prompt}")
        print("-" * 60)
        
        # 非流式请求
        response = client.chat.completions.create(
            model=model_name,
            messages=[
                {"role": "user", "content": prompt}
            ],
            max_tokens=200
        )
        
        content = response.choices[0].message.content
        usage = response.usage
        
        print(f"响应:\n{content}")
        print("-" * 60)
        print(f"Token 使用: {usage.prompt_tokens} prompt + {usage.completion_tokens} completion = {usage.total_tokens} total")
        print(f"完成原因: {response.choices[0].finish_reason}")
        print("✅ 测试成功!")
        
    except Exception as e:
        print(f"❌ 测试失败: {e}")

def test_streaming(model_name, prompt="从1数到5"):
    """测试流式响应"""
    print(f"\n{'='*60}")
    print(f"测试流式响应: {model_name}")
    print(f"{'='*60}")
    
    client = OpenAI(
        api_key=OPENBRIDGE_KEY,
        base_url=OPENBRIDGE_URL
    )
    
    try:
        print(f"发送流式请求: {prompt}")
        print("-" * 60)
        print("响应流: ", end="", flush=True)
        
        stream = client.chat.completions.create(
            model=model_name,
            messages=[
                {"role": "user", "content": prompt}
            ],
            max_tokens=100,
            stream=True
        )
        
        for chunk in stream:
            if chunk.choices[0].delta.content:
                print(chunk.choices[0].delta.content, end="", flush=True)
        
        print("\n" + "-" * 60)
        print("✅ 流式测试成功!")
        
    except Exception as e:
        print(f"\n❌ 流式测试失败: {e}")

def list_models():
    """列出所有可用模型"""
    print(f"\n{'='*60}")
    print("获取模型列表")
    print(f"{'='*60}")
    
    client = OpenAI(
        api_key=OPENBRIDGE_KEY,
        base_url=OPENBRIDGE_URL
    )
    
    try:
        models = client.models.list()
        print(f"共找到 {len(models.data)} 个模型:\n")
        
        # 按 owned_by 分组
        by_provider = {}
        for model in models.data:
            provider = model.owned_by
            if provider not in by_provider:
                by_provider[provider] = []
            by_provider[provider].append(model.id)
        
        for provider, model_list in sorted(by_provider.items()):
            print(f"Provider: {provider}")
            for model_id in sorted(model_list):
                print(f"  - {model_id}")
            print()
        
        print("✅ 获取成功!")
        
    except Exception as e:
        print(f"❌ 获取失败: {e}")

def main():
    print("""
    ╔══════════════════════════════════════════════════════════╗
    ║           OpenBridge Provider 测试脚本                  ║
    ║                                                          ║
    ║  测试不同 Provider 的功能和格式转换                      ║
    ╚══════════════════════════════════════════════════════════╝
    """)
    
    # 列出所有模型
    list_models()
    
    input("\n按 Enter 继续测试 Claude 模型...")
    
    # 测试 Claude (通过 OpenAI 兼容代理)
    test_provider(
        "claude-sonnet-4-5",
        "你好！请用一句话介绍你自己。"
    )
    
    input("\n按 Enter 继续测试流式响应...")
    
    # 测试流式响应
    test_streaming(
        "claude-sonnet-4-5",
        "请从1数到10，每个数字单独一行。"
    )
    
    print("\n" + "="*60)
    print("所有测试完成！")
    print("="*60)
    
    print("""
    💡 提示:
    
    1. 要测试 Claude 原生 API:
       - 在 config.yaml 中添加 Claude provider (type: anthropic)
       - 配置真实的 Claude API key
       - 更新路由: claude-*: claude
    
    2. 要测试 Google Gemini:
       - 在 config.yaml 中添加 Gemini provider (type: google)
       - 配置真实的 Google API key
       - 更新路由: gemini-*: gemini
       - 使用模型: gemini-1.5-pro, gemini-1.5-flash 等
    
    3. OpenBridge 会自动进行格式转换，对客户端完全透明！
    """)

if __name__ == "__main__":
    main()

