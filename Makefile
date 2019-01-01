tests_dir := test

test := ${tests_dir}/tech-db-forum
browser := firefox

api_swagger_url := https://tech-db-forum.bozaro.ru/
start_url := http://localhost:5000/api/
func_report := ${tests_dir}/func_report.html
dockerfile := Dockerfile

docker_name := docker-forum-tp
docker_tag := 1.0
docker_full_name := ${docker_name}:${docker_tag}
container_name := forum-tp



func-test:
	./${test} func --wait=30 --keep -u ${start_url} -r ${func_report}

func-test-no-keep:
	./${test} func --wait=50 -u ${start_url} -r ${func_report}

fill-test:
	./tests/tech-db-forum fill --timeout=900

perform-test:
	./tests/tech-db-forum perf --duration=600 --step=60


tests: func-test-no-keep fill-test perform-test


#--------------------------------------------------------------------------------------------------------------------------------
show-report:
	${browser} ${func_report} ${api_swagger_url} & echo "show functional test report"

clear:
	rm -rf vendor

#--------------------------------------------------------------------------------------------------------------------------------
docker-no-cache:
	docker build --no-cache -t ${docker_full_name} -f ${dockerfile} ./

docker:
	docker build -t ${docker_full_name} -f ${dockerfile} ./

	
run:
	docker run --memory 1G --log-opt max-size=1M --log-opt max-file=3 -p 5000:5000 --rm -d -it --name ${container_name} ${docker_full_name}

run-no-d:
	docker run --memory 1G --log-opt max-size=1M --log-opt max-file=3 -p 5000:5000 --rm -it --name ${container_name} ${docker_full_name}


inside:
	docker exec -it ${container_name} bash

stop:
	docker stop ${container_name}

logs:
	docker logs ${container_name}


delete-container:
	docker images
	docker rmi ${docker_full_name}
	docker images
