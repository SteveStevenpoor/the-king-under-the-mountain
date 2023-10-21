# 
RUNNABLE = <ADD YOUR RUNNABLE HERE>

# Example if you use Python:
#RUNNABLE = python greedy_dwarfs.py

# Example if you use C (do not forget to uncomment building step below):
# RUNNABLE = ./main

IN = $(wildcard tests/*-input.txt)
ACT = $(IN:-input.txt=-actual.txt)
PASS = $(IN:-input.txt=.passed)


all: clean test

clean:
	@rm -f $(PASS)
	rm -f $(ACT) $(EXE)
#	rm -rf __pycache__


# Uncomment this if you use C:

#HEADERS = $(wildcard ./*.h)

#$(RUNNABLE): *.c $(HEADERS)
#	$(CC) *.c -o $@


test: $(PASS)
	@echo "All tests passed"

$(PASS): %.passed: %-input.txt %-expected.txt # $(RUNNABLE) <--- uncomment if you use C
	@echo "Running test $*..."
	@rm -f $@
	$(RUNNABLE) $*-input.txt $*-actual.txt
	diff $*-expected.txt $*-actual.txt -Z
	@touch $@