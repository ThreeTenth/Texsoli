# 通过 readlink 获取绝对路径，再取出目录
work_path=$(dirname $(readlink -f $0))

systemctl stop saynice-api

mv $work_path/texsoli-api-amd64-linux-v0 /usr/local/bin/texsoli-api
mv $work_path/texsoli-api.service /lib/systemd/system

rm -rf $work_path

systemctl daemon-reload
systemctl enable texsoli-api

systemctl start texsoli-api