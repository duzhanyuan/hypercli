package main 

import (
	"os/exec"
	"strings"

	"github.com/docker/docker/pkg/integration/checker"
	"github.com/go-check/check"
)

func (s *DockerSuite) TestSnapshotCliCreate(c *check.C) {
	out, _ := dockerCmd(c, "volume", "create", "--name=test")
	name := strings.TrimSpace(out)
	c.Assert(name, check.Equals, "test")

	out, _ = dockerCmd(c, "snapshot", "create", "--volume=test", "--name=test-snap")
	name = strings.TrimSpace(out)
	c.Assert(name, check.Equals, "test-snap")

	out, _, err := dockerCmdWithError("snapshot", "create", "--volume=test", "--name=test-snap")
	c.Assert(err, checker.NotNil)
	c.Assert(out, checker.Contains, "conflict snapshot name(test-snap) is already assigned")
}

func (s *DockerSuite) TestSnapshotCliInspect(c *check.C) {
	c.Assert(
		exec.Command(dockerBinary, "snapshot", "inspect", "doesntexist").Run(),
		check.Not(check.IsNil),
		check.Commentf("snapshot inspect should error on non-existent volume"),
	)

	out, _ := dockerCmd(c, "volume", "create", "--name=test")
	name := strings.TrimSpace(out)
	c.Assert(name, check.Equals, "test")

	out, _ = dockerCmd(c, "snapshot", "create", "--volume=test")
	name = strings.TrimSpace(out)
	out, _ = dockerCmd(c, "snapshot", "inspect", "--format='{{ .Name }}'", name)
	c.Assert(strings.TrimSpace(out), check.Equals, name)

	dockerCmd(c, "snapshot", "create", "--volume=test", "--name=test-snap")
	out, _ = dockerCmd(c, "snapshot", "inspect", "--format='{{ .Name }}'", "test-snap")
	c.Assert(strings.TrimSpace(out), check.Equals, "test-snap")
}

func (s *DockerSuite) TestSnapshotCliInspectMulti(c *check.C) {
	out, _ := dockerCmd(c, "volume", "create", "--name=test")
	name := strings.TrimSpace(out)
	c.Assert(name, check.Equals, "test")

	dockerCmd(c, "snapshot", "create", "--volume=test", "--name=test-snap1")
	dockerCmd(c, "snapshot", "create", "--volume=test", "--name=test-snap2")
	dockerCmd(c, "snapshot", "create", "--volume=test", "--name=not-shown")	

	out, _, err := dockerCmdWithError("snapshot", "inspect", "--format='{{ .Name }}'", "test-snap1", "test-snap2", "doesntexist", "not-shown")
	c.Assert(err, checker.NotNil)
	outArr := strings.Split(strings.TrimSpace(out), "\n")
	c.Assert(len(outArr), check.Equals, 3, check.Commentf("\n%s", out))

	c.Assert(out, checker.Contains, "test-snap1")
	c.Assert(out, checker.Contains, "test-snap2")
	c.Assert(out, checker.Contains, "Error: No such snapshot: doesntexist")
	c.Assert(out, checker.Not(checker.Contains), "not-shown")
}

func (s *DockerSuite) TestSnapshotCliLs(c *check.C) {
	out, _ := dockerCmd(c, "volume", "create", "--name=test")
	name := strings.TrimSpace(out)
	c.Assert(name, check.Equals, "test")

	out, _ = dockerCmd(c, "snapshot", "create", "--volume=test")	
	id := strings.TrimSpace(out)

	dockerCmd(c, "snapshot", "create", "--volume=test", "--name=test-snap")

	out, _ = dockerCmd(c, "snapshot", "ls")
	outArr := strings.Split(strings.TrimSpace(out), "\n")
	c.Assert(len(outArr), check.Equals, 3, check.Commentf("\n%s", out))

	// Since there is no guarantee of ordering of volumes, we just make sure the names are in the output
	c.Assert(strings.Contains(out, id), check.Equals, true)
	c.Assert(strings.Contains(out, "test-snap"), check.Equals, true)
}

func (s *DockerSuite) TestSnapshotCliRm(c *check.C) {
	out, _ := dockerCmd(c, "volume", "create", "--name=test")
	name := strings.TrimSpace(out)
	c.Assert(name, check.Equals, "test")

	out, _ = dockerCmd(c, "snapshot", "create", "--volume=test")
	id := strings.TrimSpace(out)	

	dockerCmd(c, "snapshot", "create", "--volume=test", "--name", "test-snap")
	dockerCmd(c, "snapshot", "rm", id)
	dockerCmd(c, "snapshot", "rm", "test-snap")

	out, _ = dockerCmd(c, "snapshot", "ls")
	outArr := strings.Split(strings.TrimSpace(out), "\n")
	c.Assert(len(outArr), check.Equals, 1, check.Commentf("%s\n", out))

	c.Assert(
		exec.Command("snapshot", "rm", "doesntexist").Run(),
		check.Not(check.IsNil),
		check.Commentf("snapshot rm should fail with non-existent snapshot"),
	)
}

func (s *DockerSuite) TestSnapshotCliNoArgs(c *check.C) {
	out, _ := dockerCmd(c, "snapshot")
	// no args should produce the cmd usage output
	usage := "Usage:	hyper snapshot [OPTIONS] [COMMAND]"
	c.Assert(out, checker.Contains, usage)

	// invalid arg should error and show the command on stderr
	_, stderr, _, err := runCommandWithStdoutStderr(exec.Command(dockerBinary, "snapshot", "somearg"))
	c.Assert(err, check.NotNil, check.Commentf(stderr))
	c.Assert(stderr, checker.Contains, usage)

	// invalid flag should error and show the flag error and cmd usage
	_, stderr, _, err = runCommandWithStdoutStderr(exec.Command(dockerBinary, "snapshot", "--no-such-flag"))
	c.Assert(err, check.NotNil, check.Commentf(stderr))
	c.Assert(stderr, checker.Contains, usage)
	c.Assert(stderr, checker.Contains, "flag provided but not defined: --no-such-flag")
}

func (s *DockerSuite) TestSnapshotCliInspectTmplError(c *check.C) {
	out, _ := dockerCmd(c, "volume", "create", "--name=test")
	name := strings.TrimSpace(out)
	c.Assert(name, check.Equals, "test")

	out, _ = dockerCmd(c, "snapshot", "create", "--volume=test")
	name = strings.TrimSpace(out)

	out, exitCode, err := dockerCmdWithError("snapshot", "inspect", "--format='{{ .FooBar}}'", name)
	c.Assert(err, checker.NotNil, check.Commentf("Output: %s", out))
	c.Assert(exitCode, checker.Equals, 1, check.Commentf("Output: %s", out))
	c.Assert(out, checker.Contains, "Template parsing error")
}

func (s *DockerSuite) TestSnapshotCreateVol(c *check.C) {
	out, _ := dockerCmd(c, "volume", "create", "--name=test")
	name := strings.TrimSpace(out)
	c.Assert(name, check.Equals, "test")

	dockerCmd(c, "snapshot", "create", "--volume=test", "--name", "test-snap")

	dockerCmd(c, "volume", "create", "--name=snap-vol", "--snapshot=test-snap")
	out, _ = dockerCmd(c, "volume", "ls")
	c.Assert(strings.Contains(out, "snap-vol"), check.Equals, true)

	// delete, in the order snapshot, volume, volume
	out, _ = dockerCmd(c, "snapshot", "rm", "test-snap")
	name = strings.TrimSpace(out)
	c.Assert(name, check.Equals, "test-snap")

	out, _ = dockerCmd(c, "volume", "rm", "test")
	name = strings.TrimSpace(out)
	c.Assert(name, check.Equals, "test")

	out, _ = dockerCmd(c, "volume", "rm", "snap-vol")
	name = strings.TrimSpace(out)
	c.Assert(name, check.Equals, "snap-vol")
}

func (s *DockerSuite) TestSnapshotRmBasedVol(c *check.C) {
	out, _ := dockerCmd(c, "volume", "create", "--name=test")
	name := strings.TrimSpace(out)
	c.Assert(name, check.Equals, "test")

	dockerCmd(c, "snapshot", "create", "--volume=test", "--name", "test-snap")

	out, _, err := dockerCmdWithError("volume", "rm", "test")
	c.Assert(err, checker.NotNil)	
	c.Assert(out, checker.Contains, "Invalid volume: Volume still has 1 dependent snapshots")
}
