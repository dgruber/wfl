package slurmcli

import (
	"errors"
	"strings"
)

func parsesbatch(out []byte) (string, error) {
	if out == nil {
		return "", errors.New("output is nil")
	}
	// sbatch --parsable -A default ./sleep.sh
	// 16
	// --parsable: "outputs only the jobid and cluster name (if present),
	// separated by semicolon, only on successful submission."
	elements := strings.Split(string(out), ";")
	return elements[0], nil
}

func parsesqueue(out []byte) ([]string, error) {
	return []string{}, nil
}

func parsesacct(out []byte) ([]string, error) {
	/* $ sacct -P
	JobID|JobName|Partition|Account|AllocCPUS|State|ExitCode
	15|sleep.sh|debug|default|2|COMPLETED|0:0
	15.batch|batch||default|2|COMPLETED|0:0
	16|sleep.sh|debug|default|2|COMPLETED|0:0
	16.batch|batch||default|2|COMPLETED|0:0
	*/
	return []string{}, nil
}
