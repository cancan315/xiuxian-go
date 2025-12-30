#!/bin/bash
cd /root/lingma/xiuxian-go/server-go
echo "开始编译检查..."
go build -v ./cmd/server/ 2>&1
if [ $? -eq 0 ]; then
  echo "✓ 编译成功！"
else
  echo "✗ 编译失败！"
  exit 1
fi
