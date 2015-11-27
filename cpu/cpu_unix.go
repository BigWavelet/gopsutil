// +build linux freebsd darwin

package cpu

import "time"

func init() {
	lastCPUTimes, _ = CPUTimes(false)
	lastPerCPUTimes, _ = CPUTimes(true)
}

func CPUPercent(interval time.Duration, percpu bool) ([]float64, error) {
	getAllBusy := func(t CPUTimesStat) (uint64, uint64) {
		// Fields existing on kernels >= 2.6
		// (and RHEL's patched kernel 2.4...)
		// Guest time is already accounted in usertime
		user := t.User - t.Guest
		nice := t.Nice - t.GuestNice

		systemAll := t.System + t.Iowait + t.Irq + t.Softirq
		idleAll := t.Idle + t.Iowait
		virtAll := t.Guest + t.GuestNice

		total := user + nice + systemAll + idleAll + t.Steal + virtAll
		//busy := t.User + t.Nice + t.System + t.Iowait + t.Irq +
		//	t.Softirq + t.Steal + t.Guest + t.GuestNice + t.Stolen
		return total, total - idleAll //busy + t.Idle, busy
	}

	calculate := func(t1, t2 CPUTimesStat) float64 {
		t1All, t1Busy := getAllBusy(t1)
		t2All, t2Busy := getAllBusy(t2)

		if t2Busy <= t1Busy {
			return 0
		}
		if t2All <= t1All {
			return 1
		}
		return float64((t2Busy-t1Busy)*100) / float64((t2All - t1All))
	}

	cpuTimes, err := CPUTimes(percpu)
	if err != nil {
		return nil, err
	}

	if interval > 0 {
		if !percpu {
			lastCPUTimes = cpuTimes
		} else {
			lastPerCPUTimes = cpuTimes
		}
		time.Sleep(interval)
		cpuTimes, err = CPUTimes(percpu)
		if err != nil {
			return nil, err
		}
	}

	ret := make([]float64, len(cpuTimes))
	if !percpu {
		ret[0] = calculate(lastCPUTimes[0], cpuTimes[0])
		lastCPUTimes = cpuTimes
	} else {
		for i, t := range cpuTimes {
			ret[i] = calculate(lastPerCPUTimes[i], t)
		}
		lastPerCPUTimes = cpuTimes
	}
	return ret, nil
}
