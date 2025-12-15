#!/bin/sh
# SublinkPro 安装/更新脚本
# 支持全新安装、更新程序、重新安装等操作

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 安装目录
INSTALL_DIR="/usr/local/bin/sublink"
BINARY_PATH="$INSTALL_DIR/sublink"

# 打印带颜色的消息
print_info() {
    printf "${BLUE}[信息]${NC} %s\n" "$1"
}

print_success() {
    printf "${GREEN}[成功]${NC} %s\n" "$1"
}

print_warning() {
    printf "${YELLOW}[警告]${NC} %s\n" "$1"
}

print_error() {
    printf "${RED}[错误]${NC} %s\n" "$1"
}

# 检查用户是否为root
check_root() {
    if [ "$(id -u)" != "0" ]; then
        print_error "该脚本必须以root身份运行。"
        exit 1
    fi
}

# 检测是否为 Alpine
detect_alpine() {
    if [ -f /etc/alpine-release ]; then
        is_alpine=true
        # Alpine 安装依赖
        if ! command -v curl >/dev/null 2>&1; then
            print_info "正在安装依赖..."
            apk add --no-cache curl openrc libc6-compat
        fi
    else
        is_alpine=false
    fi
}

# 获取当前安装版本
get_current_version() {
    if [ -f "$BINARY_PATH" ]; then
        current_version=$("$BINARY_PATH" version 2>/dev/null || echo "未知")
    else
        current_version="未安装"
    fi
}

# 获取最新版本
get_latest_version() {
    latest_release=$(curl --silent "https://api.github.com/repos/ZeroDeng01/sublinkPro/releases/latest" \
        | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$latest_release" ]; then
        print_error "无法获取最新版本信息，请检查网络连接。"
        exit 1
    fi
}

# 检测机器类型并设置文件名
detect_machine_type() {
    machine_type=$(uname -m)
    case "$machine_type" in
        x86_64) file_name="sublinkPro-linux-amd64" ;;
        aarch64) file_name="sublinkPro-linux-arm64" ;;
        *) print_error "不支持的机器类型: $machine_type"; exit 1 ;;
    esac
}

# 停止服务
stop_service() {
    print_info "正在停止服务..."
    if [ "$is_alpine" = true ]; then
        rc-service sublink stop 2>/dev/null || true
    else
        systemctl stop sublink 2>/dev/null || true
    fi
}

# 启动服务
start_service() {
    print_info "正在启动服务..."
    if [ "$is_alpine" = true ]; then
        rc-service sublink start
    else
        systemctl start sublink
    fi
}

# 重启服务
restart_service() {
    print_info "正在重启服务..."
    if [ "$is_alpine" = true ]; then
        rc-service sublink restart
    else
        systemctl restart sublink
    fi
}

# 下载并安装二进制文件
download_and_install_binary() {
    print_info "正在下载 $latest_release 版本..."
    cd ~ || exit 1
    curl -LO "https://github.com/ZeroDeng01/sublinkPro/releases/download/$latest_release/$file_name"
    
    if [ ! -f "$file_name" ]; then
        print_error "下载失败，请检查网络连接。"
        exit 1
    fi
    
    chmod +x "$file_name"
    
    # 确保目录存在
    if [ ! -d "$INSTALL_DIR" ]; then
        mkdir -p "$INSTALL_DIR"
    fi
    
    mv "$file_name" "$BINARY_PATH"
    print_success "二进制文件安装完成。"
}

# 创建服务
create_service() {
    print_info "正在配置系统服务..."
    if [ "$is_alpine" = true ]; then
        # OpenRC 服务
        cat > /etc/init.d/sublink <<EOF
#!/sbin/openrc-run
name="sublink"
command="$INSTALL_DIR/sublink"
command_background="yes"
pidfile="/var/run/\$RC_SVCNAME.pid"
EOF
        chmod +x /etc/init.d/sublink
        rc-update add sublink default
    else
        # systemd 服务
        cat > /etc/systemd/system/sublink.service <<EOF
[Unit]
Description=Sublink Service

[Service]
ExecStart=$INSTALL_DIR/sublink
WorkingDirectory=$INSTALL_DIR
[Install]
WantedBy=multi-user.target
EOF
        systemctl daemon-reload
        systemctl enable sublink
    fi
    print_success "系统服务配置完成。"
}

# 初始化程序（设置默认密码）
init_program() {
    print_info "正在初始化程序..."
    cd "$INSTALL_DIR" || exit 1
    ./sublink setting --username admin --password 123456
}

# 清理数据目录
clean_data() {
    print_warning "正在清理数据目录..."
    rm -rf "$INSTALL_DIR/db" 2>/dev/null
    rm -rf "$INSTALL_DIR/logs" 2>/dev/null
    rm -rf "$INSTALL_DIR/template" 2>/dev/null
    print_success "数据目录已清理。"
}

# 更新程序（保留数据）
update_program() {
    print_info "========== 更新程序 =========="
    print_info "当前版本: $current_version"
    print_info "最新版本: $latest_release"
    
    stop_service
    download_and_install_binary
    start_service
    sleep 2
    restart_service
    
    print_success "========== 更新完成 =========="
    print_success "程序已更新到 $latest_release 版本"
    print_info "数据目录已保留，无需重新配置。"
}

# 全新安装
fresh_install() {
    print_info "========== 全新安装 =========="
    print_info "最新版本: $latest_release"
    
    download_and_install_binary
    create_service
    init_program
    start_service
    sleep 3
    restart_service  # workaround 首次运行是初始化，需要restart
    
    print_success "========== 安装完成 =========="
    printf "\n"
    print_success "服务已启动并设置为开机启动"
    print_info "默认账号: admin"
    print_info "默认密码: 123456"
    print_info "默认端口: 8000"
    print_warning "请登录后立即修改默认密码！"
}

# 重新安装（可选择是否保留数据）
reinstall_program() {
    print_info "========== 重新安装 =========="
    
    printf "\n是否保留现有数据？\n"
    printf "  ${GREEN}1)${NC} 保留数据（仅重装程序和服务）\n"
    printf "  ${RED}2)${NC} 清空数据（完全重新安装）\n"
    printf "请选择 [1-2]: "
    read -r data_choice
    
    stop_service
    
    case "$data_choice" in
        2)
            clean_data
            download_and_install_binary
            create_service
            init_program
            start_service
            sleep 3
            restart_service
            
            print_success "========== 重新安装完成 =========="
            print_success "服务已启动并设置为开机启动"
            print_info "默认账号: admin"
            print_info "默认密码: 123456"
            print_info "默认端口: 8000"
            print_warning "请登录后立即修改默认密码！"
            ;;
        *)
            download_and_install_binary
            create_service
            start_service
            sleep 2
            restart_service
            
            print_success "========== 重新安装完成 =========="
            print_success "程序已重新安装，数据已保留。"
            ;;
    esac
}

# 显示已安装时的菜单
show_installed_menu() {
    printf "\n"
    printf "╔══════════════════════════════════════════╗\n"
    printf "║       SublinkPro 安装/更新脚本           ║\n"
    printf "╠══════════════════════════════════════════╣\n"
    printf "║  当前版本: %-29s ║\n" "$current_version"
    printf "║  最新版本: %-29s ║\n" "$latest_release"
    printf "╚══════════════════════════════════════════╝\n"
    printf "\n"
    printf "检测到已安装 SublinkPro，请选择操作：\n"
    printf "  ${GREEN}1)${NC} 更新程序（保留所有数据）\n"
    printf "  ${YELLOW}2)${NC} 重新安装（可选择是否保留数据）\n"
    printf "  ${RED}3)${NC} 取消操作\n"
    printf "\n"
    printf "请选择 [1-3]: "
    read -r choice
    
    case "$choice" in
        1)
            update_program
            ;;
        2)
            reinstall_program
            ;;
        3)
            print_info "操作已取消。"
            exit 0
            ;;
        *)
            print_error "无效的选择。"
            exit 1
            ;;
    esac
}

# 显示未安装时的菜单
show_not_installed_menu() {
    printf "\n"
    printf "╔══════════════════════════════════════════╗\n"
    printf "║       SublinkPro 安装脚本                ║\n"
    printf "╠══════════════════════════════════════════╣\n"
    printf "║  最新版本: %-29s ║\n" "$latest_release"
    printf "╚══════════════════════════════════════════╝\n"
    printf "\n"
    
    # 检查是否存在旧数据
    if [ -d "$INSTALL_DIR/db" ] || [ -d "$INSTALL_DIR/logs" ] || [ -d "$INSTALL_DIR/template" ]; then
        print_warning "检测到存在旧数据目录！"
        printf "\n是否保留现有数据？\n"
        printf "  ${GREEN}1)${NC} 保留数据（恢复安装）\n"
        printf "  ${RED}2)${NC} 清空数据（全新安装）\n"
        printf "  ${YELLOW}3)${NC} 取消操作\n"
        printf "请选择 [1-3]: "
        read -r data_choice
        
        case "$data_choice" in
            1)
                download_and_install_binary
                create_service
                start_service
                sleep 2
                restart_service
                
                print_success "========== 恢复安装完成 =========="
                print_success "程序已安装，原有数据已保留。"
                ;;
            2)
                clean_data
                fresh_install
                ;;
            3)
                print_info "操作已取消。"
                exit 0
                ;;
            *)
                print_error "无效的选择。"
                exit 1
                ;;
        esac
    else
        fresh_install
    fi
}

# 主函数
main() {
    check_root
    detect_alpine
    get_latest_version
    detect_machine_type
    get_current_version
    
    # 检查是否已安装
    if [ -f "$BINARY_PATH" ]; then
        show_installed_menu
    else
        show_not_installed_menu
    fi
}

# 执行主函数
main
