
echo "" > logs.log

for i in {1..100}
do 
	go test -run 2A >> logs.log
done