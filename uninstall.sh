#!/bin/sh
# SublinkPro 一键卸载脚本
# 该脚本将完全卸载 SublinkPro 及其相关服务

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查用户是否为root
if [ "$(id -u)" != "0" ]; then
    printf "${RED}该脚本必须以root身份运行。${NC}\n"
    exit 1
fi

# 安装目录
INSTALL_DIR="/usr/local/bin/sublink"

# 检测是否为 Alpine
if [ -f /etc/alpine-release ]; then
    is_alpine=true
else
    is_alpine=false
fi

printf "${YELLOW}========================================${NC}\n"
printf "${YELLOW}       SublinkPro 卸载脚本${NC}\n"
printf "${YELLOW}========================================${NC}\n\n"

# 确认卸载
printf "${RED}警告: 此操作将卸载 SublinkPro 服务!${NC}\n"
printf "是否继续? [y/N]: "
read confirm
if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    printf "已取消卸载。\n"
    exit 0
fi

printf "\n${GREEN}[1/4] 停止服务...${NC}\n"

# 停止并禁用服务
if [ "$is_alpine" = true ]; then
    # OpenRC 服务
    if [ -f /etc/init.d/sublink ]; then
        rc-service sublink stop 2>/dev/null
        rc-update del sublink default 2>/dev/null
        printf "  ✓ 已停止 OpenRC 服务\n"
    else
        printf "  - OpenRC 服务不存在，跳过\n"
    fi
else
    # systemd 服务
    if [ -f /etc/systemd/system/sublink.service ]; then
        systemctl stop sublink 2>/dev/null
        systemctl disable sublink 2>/dev/null
        printf "  ✓ 已停止 systemd 服务\n"
    else
        printf "  - systemd 服务不存在，跳过\n"
    fi
fi

printf "\n${GREEN}[2/4] 删除服务文件...${NC}\n"

# 删除服务文件
if [ "$is_alpine" = true ]; then
    if [ -f /etc/init.d/sublink ]; then
        rm -f /etc/init.d/sublink
        printf "  ✓ 已删除 /etc/init.d/sublink\n"
    fi
else
    if [ -f /etc/systemd/system/sublink.service ]; then
        rm -f /etc/systemd/system/sublink.service
        systemctl daemon-reload
        printf "  ✓ 已删除 /etc/systemd/system/sublink.service\n"
    fi
fi

printf "\n${GREEN}[3/4] 删除程序文件...${NC}\n"

# 删除程序目录
if [ -d "$INSTALL_DIR" ]; then
    # 询问是否保留数据
    printf "${YELLOW}是否保留数据目录 (db、logs、template)?${NC}\n"
    printf "保留数据可用于后续重新安装时恢复 [Y/n]: "
    read keep_data
    
    if [ "$keep_data" = "n" ] || [ "$keep_data" = "N" ]; then
        # 完全删除
        rm -rf "$INSTALL_DIR"
        printf "  ✓ 已完全删除 $INSTALL_DIR (包含所有数据)\n"
    else
        # 保留数据目录
        if [ -f "$INSTALL_DIR/sublink" ]; then
            rm -f "$INSTALL_DIR/sublink"
            printf "  ✓ 已删除程序文件 $INSTALL_DIR/sublink\n"
        fi
        printf "  ✓ 已保留数据目录: $INSTALL_DIR/db, $INSTALL_DIR/logs, $INSTALL_DIR/template\n"
    fi
else
    printf "  - 程序目录不存在，跳过\n"
fi

printf "\n${GREEN}[4/4] 清理其他文件...${NC}\n"

# 删除菜单命令（如果存在）
if [ -f /usr/bin/sublink ]; then
    rm -f /usr/bin/sublink
    printf "  ✓ 已删除 /usr/bin/sublink\n"
else
    printf "  - 菜单命令不存在，跳过\n"
fi

printf "\n${GREEN}========================================${NC}\n"
printf "${GREEN}       卸载完成！${NC}\n"
printf "${GREEN}========================================${NC}\n\n"

if [ "$keep_data" != "n" ] && [ "$keep_data" != "N" ] && [ -d "$INSTALL_DIR" ]; then
    printf "${YELLOW}提示: 数据目录已保留在 $INSTALL_DIR${NC}\n"
    printf "${YELLOW}如需完全清理，请手动执行: rm -rf $INSTALL_DIR${NC}\n\n"
fi

printf "感谢使用 SublinkPro！\n"
