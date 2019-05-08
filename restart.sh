pName='tower'
towerPid=`ps -ef | grep ./$pName | grep -v grep  | awk '{print $2}'`
kill -9 $towerPid
echo "结束{$pName}进程.."
sleep 1
nohup ./${pName} > ${pName}.log  2>&1 &
echo "开启{$pName}进程.."
