package compose

import "testing"

func TestParsePorts(t *testing.T) {
	file, err := Parse([]byte(`
services:
  api:
    ports:
      - "8080:80"
      - "127.0.0.1:3306:3306"
      - target: 6379
        published: 63790
`))
	if err != nil {
		t.Fatal(err)
	}
	ports := file.Services["api"].Ports
	if len(ports) != 3 {
		t.Fatalf("expected 3 ports, got %d", len(ports))
	}
	assertPort(t, ports[0], 8080, 80, "")
	assertPort(t, ports[1], 3306, 3306, "127.0.0.1")
	assertPort(t, ports[2], 63790, 6379, "")
}

func assertPort(t *testing.T, port Port, published int, target int, hostIP string) {
	t.Helper()
	if port.Published != published || port.Target != target || port.HostIP != hostIP {
		t.Fatalf("unexpected port: %#v", port)
	}
}
