IMAGE = http1server:latest
CLUSTER = kind

build:
	docker build -t $(IMAGE) .

cluster:
	kind create cluster --name $(CLUSTER) --config kind-config.yaml

load:
	kind load docker-image $(IMAGE) --name $(CLUSTER)

deploy:
	kubectl apply -f deployment.yaml
	kubectl apply -f service.yaml
	kubectl apply -f ingress.yaml

clean:
	kubectl delete -f deployment.yaml --ignore-not-found
	kubectl delete -f service.yaml --ignore-not-found
	kubectl delete -f ingress.yaml --ignore-not-found

cluster-clean:
	kind delete cluster --name $(CLUSTER)

deploy-local: build load deploy