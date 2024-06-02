kill $(ps aux | grep [g]lfs | awk '{print $2}')
kill $(ps aux | grep [m]onitor | awk '{print $2}')
kill $(ps aux | grep [f]ail_master | awk '{print $2}')

# kill master only
# kill $(ps aux | grep '[g]lfs -role master' | awk '{print $2}')

# kill chunk server with id 1
# kill $(ps aux | grep '[g]lfs -role chunk -id 1' | awk '{print $2}')
