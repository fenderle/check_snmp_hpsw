package cmd

import (
	"fmt"
	"os"

	"github.com/olorin/nagiosplugin"
	"github.com/soniah/gosnmp"
	"github.com/spf13/cobra"
)

var (
	snmpHost         string
	snmpVersion      string
	snmpContext      string
	snmpSecLevel     string
	snmpAuthProto    string
	snmpPrivProto    string
	snmpCommunity    string
	snmpSecName      string
	snmpAuthPassword string
	snmpPrivPassword string
)

// Execute the root command setup.
func Execute() {
	cmd := &cobra.Command{
		Use:   "check_snmp_hpsw",
		Short: "Checks HP Procurve switches using SNMP",
		Long:  `Long description`,
	}

	// add persistent flags
	cmd.PersistentFlags().StringVarP(&snmpCommunity, "snmp-community", "C", "public", "SNMP Community")
	cmd.PersistentFlags().StringVarP(&snmpHost, "snmp-host", "H", "", "SNMP Host (required)")
	cmd.PersistentFlags().StringVarP(&snmpVersion, "snmp-version", "P", "2c", "SNMP Version (1|2c|3)")
	cmd.PersistentFlags().StringVarP(&snmpContext, "snmp-context", "N", "", "SNMPv3 Context")
	cmd.PersistentFlags().StringVarP(&snmpSecLevel, "snmp-security-level", "L", "noAuthNoPriv", "SNMPv3 Security Level")
	cmd.PersistentFlags().StringVarP(&snmpAuthProto, "snmp-auth-protocol", "a", "MD5", "SNMPv3 Authentication Protocol (MD5|SHA)")
	cmd.PersistentFlags().StringVarP(&snmpPrivProto, "snmp-priv-protocol", "x", "DES", "SNMPv3 Privacy Protocol (DES|AES)")
	cmd.PersistentFlags().StringVarP(&snmpSecName, "snmp-sec-name", "U", "", "SNMPv3 Username")
	cmd.PersistentFlags().StringVarP(&snmpAuthPassword, "snmp-auth-password", "A", "", "SNMPv3 Authentication Password")
	cmd.PersistentFlags().StringVarP(&snmpPrivPassword, "snmp-priv-password", "X", "", "SNMPv3 Privacy Password")

	cmd.MarkPersistentFlagRequired("snmp-host") // nolint: gas

	// add subcommands
	cmd.AddCommand(cpuCreateCommand())
	cmd.AddCommand(createMemoryCommand())
	cmd.AddCommand(createSensorCommand())

	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// parse snmp args
		err := parseSNMP()
		if err != nil {
			nagiosplugin.Exit(nagiosplugin.CRITICAL, err.Error())
		}

		// connect the SNMP host
		err = gosnmp.Default.Connect()
		if err != nil {
			nagiosplugin.Exit(nagiosplugin.CRITICAL, err.Error())
		}
	}

	cmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		// close the connection to the SNMP host
		gosnmp.Default.Conn.Close()
	}

	// run the requested command
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Parse the snmp parameters into the gosnmp.Default struct and check that
// the options are valid.
func parseSNMP() error {
	if snmpHost == "" {
		return fmt.Errorf("snmp-host is required")
	}
	gosnmp.Default.Target = snmpHost

	switch snmpVersion {
	case "1":
		return parseSNMPv1()
	case "2c":
		return parseSNMPv2c()
	case "3":
		return parseSNMPv3()
	default:
		return fmt.Errorf("snmp-version is invalid")
	}
}

// Parse the SNMPv1 options.
func parseSNMPv1() error {
	gosnmp.Default.Version = gosnmp.Version1
	return nil
}

// Parse the SNMPv2c options.
func parseSNMPv2c() error {
	gosnmp.Default.Version = gosnmp.Version2c
	gosnmp.Default.Community = snmpCommunity
	return nil
}

// Parse the SNMPv3 options.
func parseSNMPv3() error {
	gosnmp.Default.Version = gosnmp.Version3
	gosnmp.Default.ContextName = snmpContext

	// check mandatory arguments
	if snmpSecName == "" {
		return fmt.Errorf("snmp-sec-name is required")
	}

	// setup security model
	usm := &gosnmp.UsmSecurityParameters{
		UserName: snmpSecName,
	}

	gosnmp.Default.SecurityModel = gosnmp.UserSecurityModel
	gosnmp.Default.SecurityParameters = usm

	// init seclevel accordingly
	switch snmpSecLevel {
	case "noAuthNoPriv":
		gosnmp.Default.MsgFlags = gosnmp.NoAuthNoPriv
		break
	case "authNoPriv":
		gosnmp.Default.MsgFlags = gosnmp.AuthNoPriv
		if err := parseAuthProto(usm); err != nil {
			return err
		}
		fallthrough
	case "authPriv":
		gosnmp.Default.MsgFlags = gosnmp.AuthPriv
		if err := parsePrivProto(usm); err != nil {
			return err
		}
		break
	default:
		return fmt.Errorf("snmp-sec-level is invalid")
	}

	return nil
}

// Parse the AuthenticationProtocol options into the usm struct.
func parseAuthProto(usm *gosnmp.UsmSecurityParameters) error {
	if snmpAuthPassword == "" {
		return fmt.Errorf("snmp-auth-password is required")
	}
	usm.AuthenticationPassphrase = snmpAuthPassword

	switch snmpAuthProto {
	case "MD5":
		usm.AuthenticationProtocol = gosnmp.MD5
		break
	case "SHA":
		usm.AuthenticationProtocol = gosnmp.SHA
		break
	default:
		return fmt.Errorf("snmp-auth-protocol is invalid")
	}

	return nil
}

// Parse the PrivacyProtocol options into the usm struct.
func parsePrivProto(usm *gosnmp.UsmSecurityParameters) error {
	if snmpPrivPassword == "" {
		return fmt.Errorf("snmp-priv-password is required")
	}
	usm.PrivacyPassphrase = snmpPrivPassword

	switch snmpPrivProto {
	case "DES":
		usm.PrivacyProtocol = gosnmp.DES
		break
	case "AES":
		usm.PrivacyProtocol = gosnmp.AES
		break
	default:
		return fmt.Errorf("snmp-priv-protocol is invalid")
	}

	return nil
}
