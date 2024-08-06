all: cmd

cmd: pb
	docker build --build-arg svrName=discovery -t mini-router/discovery-svr .
	docker build --build-arg svrName=health -t mini-router/health-svr .
	docker build --build-arg svrName=register -t mini-router/register-svr .

pb:
	$(MAKE) -C $@

clean:
	$(MAKE) -C proto $@

.PHONY: pb