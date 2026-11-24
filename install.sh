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
RELEASES_API="https://api.github.com/repos/ZeroDeng01/sublinkPro/releases"
RELEASES_URL="https://github.com/ZeroDeng01/sublinkPro/releases"
downloaded_binary=''

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

# 筛选同页已发布且提供当前架构安装包的预发布。
preview_page() {
    LC_ALL=C awk -v asset="$file_name" '
        # Read JSON tokens so formatted release notes and nested fields
        # cannot be mistaken for release or asset metadata.
        function next_token(    result) {
            while (line !~ /[^ \t\r]/) {
                result = getline line
                if (result < 0) exit 1
                if (result == 0) {
                    token = ""
                    return
                }
            }
            sub(/^[ \t\r]+/, "", line)
            if (!match(line, /^("([^"\\[:cntrl:]]|\\(["\\\/bfnrt]|u[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]))*"|true|false|null|-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?|[][{}:,])/))
                exit 1
            token = substr(line, 1, RLENGTH)
            line = substr(line, RLENGTH + 1)
        }

        function expect(expected) {
            if (token != expected) exit 1
            next_token()
        }

        function consider_preview(tag, published,    parts, candidate) {
            if (tag !~ /^"preview-[0-9]+-[a-f0-9]+"$/ ||
                published !~ /^"[0-9-]+T[0-9:]+Z"$/ || length(published) != 22)
                return
            split(tag, parts, "-")
            if (length(parts[2]) != 8 || length(parts[3]) < 8 || length(parts[3]) > 41)
                return
            candidate = substr(published, 2, 20) " " substr(tag, 2, length(tag) - 2)
            if (candidate > best) best = candidate
        }

        function read_value(context,    object, closing, key, child, tag, published, draft, prerelease, name, state, has_asset) {
            if (token != "{" && token != "[") {
                if (token !~ /^("|[0-9-]|true$|false$|null$)/) exit 1
                next_token()
                return 0
            }

            object = token == "{"
            closing = object ? "}" : "]"
            next_token()
            if (token != closing) {
                while (1) {
                    child = ""
                    if (object) {
                        if (token !~ /^"/) exit 1
                        key = substr(token, 2, length(token) - 2)
                        next_token()
                        expect(":")
                    }
                    if (context == "releases") {
                        if (token != "{") exit 1
                        count++
                        child = "release"
                    } else if (context == "release") {
                        if (key == "tag_name") tag = token
                        else if (key == "published_at") published = token
                        else if (key == "draft") draft = token
                        else if (key == "prerelease") prerelease = token
                        else if (key == "assets" && token == "[") child = "assets"
                    } else if (context == "assets" && token == "{") {
                        child = "asset"
                    } else if (context == "asset") {
                        if (key == "name") name = token
                        else if (key == "state") state = token
                    }
                    if (read_value(child)) has_asset = 1
                    if (token != ",") break
                    next_token()
                }
            }
            expect(closing)
            if (context == "release" && draft == "false" && prerelease == "true" && has_asset)
                consider_preview(tag, published)
            if (context == "asset") return name == "\"" asset "\"" && state == "\"uploaded\""
            if (context == "assets") return has_asset
            return 0
        }

        BEGIN {
            next_token()
            if (token != "[") exit 1
            read_value("releases")
            if (token != "") exit 1
            print count + 0
            if (best != "") print best
        }
    '
}

get_latest_preview() {
    preview_page_number=1
    best_preview=''
    while :; do
        if ! release_page=$(curl --fail --silent --show-error --compressed --connect-timeout 15 --max-time 60 \
            "$RELEASES_API?per_page=100&page=$preview_page_number"); then
            print_error "无法获取预发布列表，请检查网络或 GitHub API 限流后重试。"
            exit 1
        fi
        if ! parsed_page=$(printf '%s\n' "$release_page" | preview_page); then
            print_error "无法解析 GitHub 发布列表。"
            exit 1
        fi

        page_length=$(printf '%s\n' "$parsed_page" | sed -n '1p')
        candidate=$(printf '%s\n' "$parsed_page" | sed -n '2p')
        if [ -n "$candidate" ]; then
            best_preview=$(printf '%s\n%s\n' "$best_preview" "$candidate" | LC_ALL=C sort -r | head -n 1)
        fi
        [ "$page_length" -eq 100 ] || break
        preview_page_number=$((preview_page_number + 1))
    done

    latest_release=${best_preview#* }
    if [ -z "$latest_release" ]; then
        print_error "没有找到适用于当前架构的预发布安装包（${file_name}）。"
        exit 1
    fi
}

# 获取所选渠道的最新版本
get_latest_version() {
    if [ "$release_channel" = 'preview' ]; then
        get_latest_preview
        return
    fi

    if ! release_json=$(curl --fail --silent --show-error --connect-timeout 15 --max-time 60 "$RELEASES_API/latest"); then
        print_error "无法获取最新稳定版，请检查网络连接。"
        exit 1
    fi
    latest_release=$(printf '%s\n' "$release_json" | sed -n -E 's/^[[:space:]]*"tag_name"[[:space:]]*:[[:space:]]*"([A-Za-z0-9._-]+)".*/\1/p')
    if [ -z "$latest_release" ]; then
        print_error "无法识别最新稳定版的标签。"
        exit 1
    fi
}

# 检测机器类型并设置文件名
detect_machine_type() {
    machine_type=$(uname -m)
    case "$machine_type" in
        x86_64) file_name="sublinkPro-linux-amd64" ;;
        aarch64) file_name="sublinkPro-linux-arm64" ;;
        armv7l|armv7*) file_name="sublinkPro-linux-armv7" ;;
        i386|i686) file_name="sublinkPro-linux-x86" ;;
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

# 下载完成后再停止服务、替换程序或清理数据。
cleanup_download() {
    if [ -n "$downloaded_binary" ]; then
        rm -f "$downloaded_binary"
    fi
}

download_binary() {
    print_info "正在下载 $latest_release 版本..."
    if ! mkdir -p "$INSTALL_DIR" || ! downloaded_binary=$(mktemp "$INSTALL_DIR/.sublink-download.XXXXXX"); then
        print_error "无法在安装目录中创建下载文件。"
        exit 1
    fi
    trap 'cleanup_download' 0
    trap 'exit 1' HUP INT TERM

    if ! curl --fail --silent --show-error --location --connect-timeout 15 \
        --output "$downloaded_binary" "$RELEASES_URL/download/$latest_release/$file_name" \
        || [ ! -s "$downloaded_binary" ] || ! chmod 755 "$downloaded_binary"; then
        print_error "无法准备 $latest_release 的安装包，请检查下载和安装目录权限。"
        exit 1
    fi
}

install_downloaded_binary() {
    if ! mv -f "$downloaded_binary" "$BINARY_PATH"; then
        print_error "无法替换二进制文件。"
        if [ -x "$BINARY_PATH" ]; then
            start_service
        fi
        exit 1
    fi
    downloaded_binary=''
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
    
    download_binary
    stop_service
    install_downloaded_binary
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
    
    download_binary
    if [ "${1:-}" = 'clear-data' ]; then
        clean_data
    fi
    install_downloaded_binary
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
    
    download_binary
    stop_service
    
    case "$data_choice" in
        2)
            clean_data
            install_downloaded_binary
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
            install_downloaded_binary
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
                download_binary
                install_downloaded_binary
                create_service
                start_service
                sleep 2
                restart_service
                
                print_success "========== 恢复安装完成 =========="
                print_success "程序已安装，原有数据已保留。"
                ;;
            2)
                fresh_install 'clear-data'
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
    release_channel='stable'
    case "${1:-}" in
        --preview) release_channel='preview'; shift ;;
        --help|-h)
            printf '%s\n' '用法: sh install.sh [--preview]' '默认安装稳定版。--preview 安装或更新到最新预发布。'
            exit 0
            ;;
    esac
    if [ "$#" -ne 0 ]; then
        print_error "不支持的参数，请使用 --help 查看用法。"
        exit 1
    fi

    check_root
    detect_alpine
    detect_machine_type
    get_latest_version
    get_current_version
    
    # 检查是否已安装
    if [ -f "$BINARY_PATH" ]; then
        show_installed_menu
    else
        show_not_installed_menu
    fi
}

# 执行主函数
main "$@"
