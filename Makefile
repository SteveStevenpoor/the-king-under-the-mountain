# 
RUNNABLE = main.go

IN = $(wildcard tests/*-input.txt)
ACT = $(IN:-input.txt=-actual.txt)
PASS = $(IN:-input.txt=.passed)


all: clean test

clean:
	@rm -f $(PASS)
	rm -f $(ACT) $(EXE)


test: $(PASS)
	@echo "All tests passed"

$(PASS): %.passed: %-input.txt %-expected.txt
	@echo "Running test $*..."
	@rm -f $@
	go run $(RUNNABLE) $*-input.txt $*-actual.txt
	diff $*-expected.txt $*-actual.txt -Z
	@touch $@

