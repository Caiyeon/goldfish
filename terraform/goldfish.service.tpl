[Unit]
Description=goldfish
Requires=network-online.target
After=network-online.target
[Service]
User=root
Group=root
WorkingDirectory=/home/ubuntu
ExecStart=/home/ubuntu/goldfish-linux-amd64 -config=/etc/goldfish-config.hcl -token=${WRAPPING_TOKEN}
ExecReload=/bin/kill -HUP $MAINPID
KillSignal=SIGINT
[Install]
WantedBy=multi-user.target
