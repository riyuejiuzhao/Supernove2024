all: proto

proto:
	$(MAKE) -C $@

clean:
	$(MAKE) -C proto $@

.PHONY: proto