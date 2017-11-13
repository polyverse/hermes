echo "Kill everything by running: killall standalone"

killall standalone

echo "Building..."
go build

echo "Starting base test app..."
./standalone -children "child3=http://localhost:9094,child4=http://localhost:9095" &

echo "Starting a child app with push..."
./standalone -serveraddr ":9092" -parenturl "http://localhost:9091" -ourname "child1" &

echo "Starting a second child app with push..."
./standalone -serveraddr ":9093" -parenturl "http://localhost:9091" -ourname "child2" &

echo "Starting a third child app without push..."
./standalone -serveraddr ":9094" &

echo "Starting a fourth child app without push..."
./standalone -serveraddr ":9095" -children "subchild2=http://localhost:9097" &

echo "Starting a grand-child...."
./standalone -serveraddr ":9097" &

echo "Starting a second grand-child..."
HERMES_PARENT_URL=http://localhost:9095 HERMES_MYNAME=subchild1 ./standalone -serveraddr ":9096" -parentfromenv=true &
