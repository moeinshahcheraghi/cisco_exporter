package login

import (
    "errors"
    "regexp"
    "strings"
    "github.com/moeinshahcheraghi/cisco_exporter/rpc"
)

func Parse(ostype string, output string) (int, error) {
    if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
        return 0, errors.New("'show login failures' is not implemented for " + ostype)
    }
    failureRegexp := regexp.MustCompile(`Username:\s+(\S+),\s+IP:\s+(\S+),\s+Time:\s+(.+)`)
    lines := strings.Split(output, "\n")
    count := 0
    for _, line := range lines {
        if failureRegexp.MatchString(line) {
            count++
        }
    }
    return count, nil
}

func ParseLogging(ostype string, output string) (int, error) {
    if ostype != rpc.IOSXE && ostype != rpc.NXOS && ostype != rpc.IOS {
        return 0, errors.New("'show logging' is not implemented for " + ostype)
    }
    loginRegexp := regexp.MustCompile(`%SEC_LOGIN-5-LOGIN_SUCCESS`)
    lines := strings.Split(output, "\n")
    count := 0
    for _, line := range lines {
        if loginRegexp.MatchString(line) {
            count++
        }
    }
    return count, nil
}