package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"shareDir/iptracker/peer"
	"shareDir/iptracker/tracker"
	"strconv"
	"syscall"
)

const envDaemonChild = "SHAREDIR_DAEMON_CHILD"

// Behavior abstracts the specific work of the daemon (Server or Client)
type Behavior interface {
	Start() error
	Stop(error)
}

// TrackerBehavior implements server-side behavior using the tracker component
type TrackerBehavior struct {
	*tracker.Tracker
}

func NewTrackerBehavior(port int) *TrackerBehavior {
	return &TrackerBehavior{tracker.New(port)}
}

func (t *TrackerBehavior) Start() error {
	return t.Tracker.Start()
}

func (t *TrackerBehavior) Stop(err error) {
	t.Tracker.Stop(err)
}

func (t *TrackerBehavior) GetPeers() []string {
	return t.Tracker.GetPeers()
}

// PeerBehavior implements client-side behavior using the peer component
type PeerBehavior struct {
	*peer.Peer
}

func NewPeerBehavior(id, addr string) *PeerBehavior {
	return &PeerBehavior{peer.New(id, addr)}
}

func (p *PeerBehavior) Start() error {
	return p.Peer.Start()
}

func (p *PeerBehavior) Stop(err error) {
	p.Peer.Stop(err)
}

// Daemon is the high-level coordinator that handles lifecycle and OS signals
type Daemon struct {
	Name     string
	Behavior Behavior
	pidFile  string
	logFile  string
}

func New(name string, behavior Behavior) *Daemon {
	return &Daemon{
		Name:     name,
		Behavior: behavior,
		pidFile:  fmt.Sprintf("/tmp/share_%s.pid", name),
		logFile:  fmt.Sprintf("/tmp/share_%s.log", name),
	}
}

// Run starts the daemon, optionally in the background
func (d *Daemon) Run(daemonize bool) error {
	if daemonize && os.Getenv(envDaemonChild) == "" {
		// Parent process: spawn child and exit
		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		cmd.Env = append(os.Environ(), envDaemonChild+"=1")

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start background process: %w", err)
		}

		fmt.Printf("[%s] Started in background [PID %d]\n", d.Name, cmd.Process.Pid)
		os.Exit(0)
	}

	// 1. PID management
	if data, err := os.ReadFile(d.pidFile); err == nil {
		pid, _ := strconv.Atoi(string(data))
		if proc, err := os.FindProcess(pid); err == nil {
			// On Unix, FindProcess always succeeds, we need to check if it's alive
			if err := proc.Signal(syscall.Signal(0)); err == nil {
				return fmt.Errorf("%s is already running [PID %d]", d.Name, pid)
			}
		}
	}

	pid := os.Getpid()
	if err := os.WriteFile(d.pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("failed to write pid file: %w", err)
	}
	defer os.Remove(d.pidFile)

	// 2. Output redirection for background process
	if os.Getenv(envDaemonChild) != "" && d.logFile != "" {
		logF, err := os.OpenFile(d.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			syscall.Dup2(int(logF.Fd()), int(os.Stdout.Fd()))
			syscall.Dup2(int(logF.Fd()), int(os.Stderr.Fd()))
			defer logF.Close()
		}
	}

	// 3. Signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		fmt.Printf("\n[%s] Received signal: %v, shutting down...\n", d.Name, sig)
		d.Behavior.Stop(fmt.Errorf("received signal %v", sig))
	}()

	fmt.Printf("[%s] Starting daemon with PID %d\n", d.Name, pid)

	// 4. Start the actual behavior (tracker or peer)
	if err := d.Behavior.Start(); err != nil {
		return fmt.Errorf("daemon %s error: %w", d.Name, err)
	}

	return nil
}

// Stop sends a signal to the running daemon process
func (d *Daemon) Stop() error {
	data, err := os.ReadFile(d.pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s is not running (pid file not found)", d.Name)
		}
		return fmt.Errorf("failed to read pid file: %w", err)
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return fmt.Errorf("invalid pid file content: %w", err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to signal process %d: %w", pid, err)
	}

	fmt.Printf("[%s] Sent stop signal to process %d\n", d.Name, pid)
	return nil
}
