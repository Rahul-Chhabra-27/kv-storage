#!/bin/bash

# ---- CONFIG ----
OUTPUT="performance_metrics.csv"
MYSQL_USER="root"
MYSQL_PASS="aec19cs047"
MYSQL_SOCKET="/tmp/mysql.sock"

# Write header once
echo "timestamp,cpu_%,disk_read_iops,disk_write_iops,mb_s,threads_running,pending_writes" > $OUTPUT

echo "Monitoring... Press CTRL+C to stop."

while true
do
    timestamp=$(date +"%Y-%m-%d %H:%M:%S")

    # CPU Utilization (%)
    cpu=$(top -l 1 | awk '/CPU usage/ {print $3}' | sed 's/%//')

    # Disk metrics: read IOPS, write IOPS, MB/s
    disk=$(iostat -d disk0 1 2 | tail -n 1)
    read_iops=$(echo $disk | awk '{print $3}')
    write_iops=$(echo $disk | awk '{print $4}')
    mb_s=$(echo $disk | awk '{print $6}')

    # MySQL internal metrics
    threads_running=$(mysql -u$MYSQL_USER -p$MYSQL_PASS -S $MYSQL_SOCKET -Nse "SHOW GLOBAL STATUS LIKE 'Threads_running';" | awk '{print $2}')
    pending_writes=$(mysql -u$MYSQL_USER -p$MYSQL_PASS -S $MYSQL_SOCKET -Nse "SHOW ENGINE INNODB STATUS" | grep -i "Pending writes" | awk '{print $3}')

    # Write row to CSV
    echo "$timestamp,$cpu,$read_iops,$write_iops,$mb_s,$threads_running,$pending_writes" >> $OUTPUT

    sleep 1
done
