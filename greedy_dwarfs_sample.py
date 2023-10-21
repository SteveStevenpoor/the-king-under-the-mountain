import sys


def solution(data: str) -> str:
    pass


def run():
    if len(sys.argv) <= 2:
        print("Wrong number of arguments: please specify input and output files.")
        return

    input_name, output_name = sys.argv[1], sys.argv[2]
    with open(input_name, "r") as inp:
        with open(output_name, "w") as out:
            data = inp.read()
            result = solution(data)
            out.write(result)


run()
