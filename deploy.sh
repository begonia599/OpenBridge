#!/bin/bash

#######################################
# OpenBridge 一键部署脚本
# 适用于 Ubuntu/Debian 系统
# 功能:
# - 自动检测 Docker 安装状态
# - 自动检测可用端口
# - 自动配置防火墙
# - 部署 OpenBridge 服务
#######################################

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 打印横幅
print_banner() {
    echo -e "${BLUE}"
    echo "╔═══════════════════════════════════════════╗"
    echo "║                                           ║"
    echo "║         OpenBridge 一键部署脚本           ║"
    echo "║                                           ║"
    echo "║     OpenAI API Gateway for AssemblyAI     ║"
    echo "║                                           ║"
    echo "╚═══════════════════════════════════════════╝"
    echo -e "${NC}"
}

# 检查是否为 root 用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "此脚本需要 root 权限运行"
        log_info "请使用: sudo $0"
        exit 1
    fi
}

# 检测系统类型
detect_system() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VER=$VERSION_ID
        log_info "检测到系统: $OS $VER"
    else
        log_error "无法检测系统类型"
        exit 1
    fi
}

# 检查 Docker 安装状态
check_docker() {
    log_info "检查 Docker 安装状态..."
    
    if command -v docker &> /dev/null; then
        DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
        log_success "Docker 已安装: $DOCKER_VERSION"
        return 0
    else
        log_warning "Docker 未安装"
        return 1
    fi
}

# 安装 Docker
install_docker() {
    log_info "开始安装 Docker..."
    
    # 更新包索引
    apt-get update
    
    # 安装依赖
    apt-get install -y \
        ca-certificates \
        curl \
        gnupg \
        lsb-release
    
    # 添加 Docker 官方 GPG key
    mkdir -p /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/$OS/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    
    # 设置仓库
    echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$OS \
        $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
    
    # 安装 Docker Engine
    apt-get update
    apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    
    # 启动 Docker
    systemctl start docker
    systemctl enable docker
    
    log_success "Docker 安装完成"
}

# 检查 Docker Compose
check_docker_compose() {
    log_info "检查 Docker Compose..."
    
    # 检查新版本 (docker compose)
    if docker compose version &> /dev/null; then
        COMPOSE_VERSION=$(docker compose version --short)
        COMPOSE_CMD="docker compose"
        log_success "Docker Compose (plugin) 已安装: $COMPOSE_VERSION"
        return 0
    # 检查旧版本 (docker-compose)
    elif command -v docker-compose &> /dev/null; then
        COMPOSE_VERSION=$(docker-compose --version | awk '{print $3}' | sed 's/,//')
        COMPOSE_CMD="docker-compose"
        log_success "Docker Compose (standalone) 已安装: $COMPOSE_VERSION"
        return 0
    else
        log_error "Docker Compose 未安装"
        return 1
    fi
}

# 检测可用端口
find_available_port() {
    local start_port=${1:-8080}
    local max_attempts=100
    local port=$start_port
    
    log_info "检测可用端口 (从 $start_port 开始)..."
    
    for ((i=0; i<max_attempts; i++)); do
        if ! netstat -tuln 2>/dev/null | grep -q ":$port " && \
           ! ss -tuln 2>/dev/null | grep -q ":$port "; then
            log_success "找到可用端口: $port"
            echo $port
            return 0
        fi
        ((port++))
    done
    
    log_error "未找到可用端口"
    exit 1
}

# 配置防火墙
configure_firewall() {
    local port=$1
    
    log_info "配置防火墙规则..."
    
    # 检查 UFW
    if command -v ufw &> /dev/null; then
        log_info "检测到 UFW 防火墙"
        
        # 检查 UFW 是否启用
        if ufw status | grep -q "Status: active"; then
            log_info "开放端口 $port (UFW)..."
            ufw allow $port/tcp
            log_success "UFW 规则已添加"
        else
            log_warning "UFW 未启用,跳过防火墙配置"
        fi
    # 检查 firewalld
    elif command -v firewall-cmd &> /dev/null; then
        log_info "检测到 firewalld 防火墙"
        
        if systemctl is-active --quiet firewalld; then
            log_info "开放端口 $port (firewalld)..."
            firewall-cmd --permanent --add-port=$port/tcp
            firewall-cmd --reload
            log_success "firewalld 规则已添加"
        else
            log_warning "firewalld 未启用,跳过防火墙配置"
        fi
    # 检查 iptables
    elif command -v iptables &> /dev/null; then
        log_info "检测到 iptables 防火墙"
        log_info "开放端口 $port (iptables)..."
        iptables -A INPUT -p tcp --dport $port -j ACCEPT
        
        # 尝试保存规则
        if command -v iptables-save &> /dev/null; then
            iptables-save > /etc/iptables/rules.v4 2>/dev/null || true
        fi
        log_success "iptables 规则已添加"
    else
        log_warning "未检测到防火墙,跳过配置"
    fi
}

# 配置 config.yaml
configure_config() {
    log_info "配置 config.yaml..."
    
    if [ ! -f config.yaml ]; then
        log_error "config.yaml 不存在"
        exit 1
    fi
    
    # 提示用户输入 API Keys
    echo ""
    log_info "请配置 AssemblyAI API Keys"
    echo -e "${YELLOW}提示: 至少需要一个 API Key${NC}"
    echo ""
    
    read -p "请输入第一个 AssemblyAI API Key: " api_key_1
    
    if [ -z "$api_key_1" ]; then
        log_error "必须提供至少一个 API Key"
        exit 1
    fi
    
    # 备份原配置
    cp config.yaml config.yaml.bak
    
    # 更新配置文件
    sed -i "s/- \"a266077175884a0abd7c5d094de90c39\"/- \"$api_key_1\"/" config.yaml
    
    # 询问是否添加更多 keys
    read -p "是否添加更多 API Keys? (y/n): " add_more
    
    if [[ $add_more == "y" || $add_more == "Y" ]]; then
        read -p "请输入第二个 API Key (直接回车跳过): " api_key_2
        if [ ! -z "$api_key_2" ]; then
            sed -i "/- \"$api_key_1\"/a\    - \"$api_key_2\"" config.yaml
        fi
        
        read -p "请输入第三个 API Key (直接回车跳过): " api_key_3
        if [ ! -z "$api_key_3" ]; then
            sed -i "/- \"$api_key_2\"/a\    - \"$api_key_3\"" config.yaml
        fi
    fi
    
    log_success "配置文件已更新"
}

# 部署服务
deploy_service() {
    local port=$1
    
    log_info "开始部署 OpenBridge..."
    
    # 设置端口环境变量
    export PORT=$port
    
    # 停止旧容器
    log_info "停止旧容器..."
    $COMPOSE_CMD down 2>/dev/null || true
    
    # 构建并启动
    log_info "构建 Docker 镜像..."
    $COMPOSE_CMD build
    
    log_info "启动服务..."
    $COMPOSE_CMD up -d
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 5
    
    # 检查服务状态
    if $COMPOSE_CMD ps | grep -q "Up"; then
        log_success "服务启动成功!"
    else
        log_error "服务启动失败"
        log_info "查看日志:"
        $COMPOSE_CMD logs
        exit 1
    fi
}

# 显示部署信息
show_deployment_info() {
    local port=$1
    local ip=$(hostname -I | awk '{print $1}')
    
    echo ""
    echo -e "${GREEN}╔═══════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                           ║${NC}"
    echo -e "${GREEN}║          🎉 部署成功! 🎉                  ║${NC}"
    echo -e "${GREEN}║                                           ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}服务信息:${NC}"
    echo -e "  • 本地访问: ${GREEN}http://localhost:$port${NC}"
    echo -e "  • 内网访问: ${GREEN}http://$ip:$port${NC}"
    echo -e "  • 外网访问: ${GREEN}http://YOUR_PUBLIC_IP:$port${NC}"
    echo ""
    echo -e "${BLUE}API 端点:${NC}"
    echo -e "  • Chat Completions: ${GREEN}/v1/chat/completions${NC}"
    echo -e "  • Models List: ${GREEN}/v1/models${NC}"
    echo -e "  • Health Check: ${GREEN}/health${NC}"
    echo -e "  • Statistics: ${GREEN}/stats${NC}"
    echo ""
    echo -e "${BLUE}常用命令:${NC}"
    echo -e "  • 查看日志: ${YELLOW}$COMPOSE_CMD logs -f${NC}"
    echo -e "  • 停止服务: ${YELLOW}$COMPOSE_CMD stop${NC}"
    echo -e "  • 启动服务: ${YELLOW}$COMPOSE_CMD start${NC}"
    echo -e "  • 重启服务: ${YELLOW}$COMPOSE_CMD restart${NC}"
    echo -e "  • 查看状态: ${YELLOW}$COMPOSE_CMD ps${NC}"
    echo ""
    echo -e "${BLUE}测试命令:${NC}"
    echo -e "${YELLOW}curl http://localhost:$port/health${NC}"
    echo ""
}

# 主函数
main() {
    print_banner
    
    # 检查 root 权限
    check_root
    
    # 检测系统
    detect_system
    
    # 检查并安装 Docker
    if ! check_docker; then
        read -p "是否安装 Docker? (y/n): " install_docker_choice
        if [[ $install_docker_choice == "y" || $install_docker_choice == "Y" ]]; then
            install_docker
        else
            log_error "Docker 是必需的,退出安装"
            exit 1
        fi
    fi
    
    # 检查 Docker Compose
    if ! check_docker_compose; then
        log_error "Docker Compose 未安装,请先安装 Docker Compose"
        exit 1
    fi
    
    # 查找可用端口
    PORT=$(find_available_port 8080)
    
    # 询问是否使用该端口
    read -p "使用端口 $PORT? (y/n, 直接回车使用): " use_port
    if [[ $use_port == "n" || $use_port == "N" ]]; then
        read -p "请输入要使用的端口: " custom_port
        PORT=$custom_port
    fi
    
    # 配置防火墙
    configure_firewall $PORT
    
    # 配置 config.yaml
    configure_config
    
    # 部署服务
    deploy_service $PORT
    
    # 显示部署信息
    show_deployment_info $PORT
    
    log_success "部署完成!"
}

# 运行主函数
main "$@"
