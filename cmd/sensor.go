package cmd

import (
	"github.com/olorin/nagiosplugin"
	"github.com/soniah/gosnmp"
	"github.com/spf13/cobra"
)

const (
	hpicfSensorStatus = "1.3.6.1.4.1.11.2.14.11.1.2.6.1.4"
	hpicfSensorDescr  = "1.3.6.1.4.1.11.2.14.11.1.2.6.1.7"
)

type sensorStatus int

const (
	sensorUnknown    sensorStatus = 1
	sensorBad        sensorStatus = 2
	sensorWarning    sensorStatus = 3
	sensorGood       sensorStatus = 4
	sensorNotPresent sensorStatus = 5
)

func createSensorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sensor",
		Short: "Check the sensors",
		Long: `Query the switch, retrieve the sensor data and check for
failures.`,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		// query the SNMP host for the descriptions
		descriptions := []string{}
		if err := gosnmp.Default.BulkWalk(hpicfSensorDescr, func(pdu gosnmp.SnmpPDU) error {
			descriptions = append(descriptions, string(pdu.Value.([]byte)))
			return nil
		}); err != nil {
			nagiosplugin.Exit(nagiosplugin.CRITICAL, err.Error())
		}

		// query the SNMP host for the states
		states := []sensorStatus{}
		if err := gosnmp.Default.BulkWalk(hpicfSensorStatus, func(pdu gosnmp.SnmpPDU) error {
			value := gosnmp.ToBigInt(pdu.Value).Uint64()
			states = append(states, sensorStatus(value))
			return nil
		}); err != nil {
			nagiosplugin.Exit(nagiosplugin.CRITICAL, err.Error())
		}

		// make sure the slices have the same size
		if len(descriptions) != len(states) {
			nagiosplugin.Exit(nagiosplugin.CRITICAL, "invalid sensordata")
		}

		// instance a new nagios check
		check := nagiosplugin.NewCheck()
		defer check.Finish()

		for idx, state := range states {
			desc := descriptions[idx]

			switch state {
			case sensorGood:
				check.AddResultf(nagiosplugin.OK, "%s: Good", desc)
				break
			case sensorWarning:
				check.AddResultf(nagiosplugin.WARNING, "%s: Warning", desc)
				break
			case sensorBad:
				check.AddResultf(nagiosplugin.CRITICAL, "%s: Error", desc)
				break
			case sensorUnknown:
				check.AddResultf(nagiosplugin.UNKNOWN, "%s: Unknown", desc)
				break
			default:
				break
			}
		}
	}

	return cmd
}
