package server

func msToTicks(ms uint, simPeriodMs float64) int {
	if simPeriodMs <= 0 {
		simPeriodMs = 125 // default simulation period (ms), 8 ticks/s
	}
	return int(float64(ms) / simPeriodMs)
}

// defaultObserverUpdateIntervalTicks is the stride when
// -observerUpdateIntervalMS is not on the command line: 8 TICKS, which is what
// the server starts with and only overwrites when the flag was given.
//
// The number is not cosmetic. The observer update strides the observer list by
// it, so 8 means every observer is served once every ninth tick: at a 125 ms
// period that is the 200 frames a second capture 2026-09-29_012528_03cb67c7cf19
// measured, in bursts of 200 at a time. Reading the flag's default of 8 as
// milliseconds instead gives 0 ticks, which is the protocol's own infinite loop
// (the stride would never advance), and the guard that saves it from hanging
// turns into 1, which serves every observer on every tick: eight visits a
// second, and a late observer catching up eight times faster than it should.
const defaultObserverUpdateIntervalTicks = 8

// observerUpdateIntervalTicks is the stride the session's observer update walks
// the observer list with.
func (s *Server) observerUpdateIntervalTicks() int {
	return ObserverUpdateIntervalTicks(s.config)
}

// ObserverUpdateIntervalTicks is the stride in TICKS for a configuration: the
// 8-tick default, or the flag converted from milliseconds when it was given.
//
// It is exported because **two** places need it and they must not disagree: the server,
// which stores it in the session, and the simulation harness, which has to know how long
// the observer step takes to come round so that its settle covers one visit
// (internal/sim/replay.go, settleFor). A settle computed from a different number than the
// one the step walks with is a settle that stops reading before the thing it is waiting
// for exists, and the harness would then report a missing frame as a difference of the
// server that is really its own.
func ObserverUpdateIntervalTicks(cfg *Config) int {
	if cfg == nil || !cfg.ObserverUpdateIntervalSet {
		return defaultObserverUpdateIntervalTicks
	}
	return msToTicks(cfg.ObserverUpdateIntervalMS, cfg.SimulationPeriod)
}

// observerSocketSkipTicks returns the observer socket processing skip in ticks,
// fed from the observerSocketSkipMS setting.
func (s *Server) observerSocketSkipTicks() uint32 {
	if s.config == nil {
		return 0
	}
	ticks := msToTicks(s.config.ObserverSocketSkipMS, s.config.SimulationPeriod)
	ticks = max(ticks, 0)
	return uint32(ticks)
}