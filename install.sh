#!/bin/sh
# 检查用户是否为root
if [ "$(id -u)" != "0" ]; then
    printf "该脚本必须以root身份运行。\n"
    exit 1
fi

# 检测是否为 Alpine
if [ -f /etc/alpine-release ]; then
    is_alpine=true
    # Alpine 
    if ! command -v curl >/dev/null 2>&1; then
        apk add --no-cache curl openrc libc6-compat
    fi
else
    is_alpine=false
fi

# 创建程序目录
INSTALL_DIR="/usr/local/bin/sublink"
if [ ! -d "$INSTALL_DIR" ]; then
    mkdir -p "$INSTALL_DIR"
fi

# 获取最新的发行版标签
latest_release=$(curl --silent "https://api.github.com/repos/ZeroDeng01/sublinkPro/releases/latest" \
    | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
printf "最新版本: %s\n" "$latest_release"

# 检测机器类型
machine_type=$(uname -m)
case "$machine_type" in
    x86_64) file_name="sublinkPro-linux-amd64" ;;
    aarch64) file_name="sublinkPro-linux-arm64" ;;
    *) printf "不支持的机器类型: %s\n" "$machine_type"; exit 1 ;;
esac

# 下载文件
cd ~ || exit 1
curl -LO "https://github.com/ZeroDeng01/sublinkPro/releases/download/$latest_release/$file_name"

# 设置可执行
chmod +x "$file_name"

# 移动到指定目录
mv "$file_name" "$INSTALL_DIR/sublink"

# 初始化系统
cd $INSTALL_DIR 
./sublink setting --username admin --password 123456

# 创建服务
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
    rc-service sublink start
    sleep 3
    rc-service sublink restart # workaround 首次运行是初始化，需要restart
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
    systemctl start sublink
    systemctl enable sublink
    sleep 3
    systemctl restart sublink # workaround 首次运行是初始化，需要restart
fi



printf "服务已启动并已设置为开机启动\n"
printf "默认账号 admin 密码 123456 默认端口 8000\n"

# TODO: support alpine in menu.sh
#printf "安装完成已经启动 输入 sublink 可以呼出菜单\n"
# 下载 menu.sh 并设置权限
#curl -o /usr/bin/sublink -H "Cache-Control: no-cache" -H "Pragma: no-cache" \
#    https://raw.githubusercontent.com/ZeroDeng01/sublinkPro/main/menu.sh
#chmod 755 "/usr/bin/sublink"
