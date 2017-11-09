echo "Kill everything by running: killall test_app"

killall test_app

echo "Building..."
go build

echo "Starting base test app..."
./test_app -children "child3=http://localhost:9094,child4=http://localhost:9095" &

echo "Starting a child app with push..."
./test_app -serveraddr ":9092" -parenturl "http://localhost:9091" -ourname "child1" &

echo "Starting a second child app with push..."
./test_app -serveraddr ":9093" -parenturl "http://localhost:9091" -ourname "child2" &

echo "Starting a third child app without push..."
./test_app -serveraddr ":9094" &

echo "Starting a fourth child app without push..."
./test_app -serveraddr ":9095" -children "subchild2=http://localhost:9097" &

echo "Starting a grand-child...."
./test_app -serveraddr ":9096" --parenturl "http://localhost:9095" -ourname "subchild1" &

echo "Starting a second grand-child..."
./test_app -serveraddr ":9097"
