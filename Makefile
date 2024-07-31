all: pb

pb:
	$(MAKE) -C $@

clean:
	$(MAKE) -C proto $@

.PHONY: pb