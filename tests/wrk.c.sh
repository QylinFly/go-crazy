clear
wrk -t2 -c150 -d20s -T5 --script=./post.lua --latency http://10.99.2.116:8087/invoke
