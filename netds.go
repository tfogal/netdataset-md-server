// Hacky program to echo our hacky iv3d protocol.
package main;
import "bufio"
import "encoding/binary"
import "fmt"
import "io"
import "net"
import "os"

const(
  port = ":4445"
)

func main() {
  lstn, err := net.Listen("tcp", port)
  if err != nil {
    fmt.Fprintf(os.Stderr, "could not start create listen socket: %s\n", err)
    return
  }
  fmt.Printf("Startup complete, waiting for connections on localhost%s\n", port)
  for {
    cnxn, err := lstn.Accept()
    if err != nil {
      fmt.Fprintf(os.Stderr, "error accepting client connection: %s\n", err)
    }
    go echoCmds(cnxn)
  }
}

const(
  cmd_OPEN=0
  cmd_CLOSE=iota
  cmd_BRICK=iota
)

// checks for magic on the given stream.
func magic(rdr *bufio.Reader) (error) {
  data := make([]byte, 4)
  n, err := rdr.Read(data[0:4])
  if err != nil { return err }
  if n != 4 { return err }
  if data[0] != 'I' || data[1] != 'V' || data[2] != '3' || data[3] != 'D' {
    return err
  }
  return nil
}

func echoCmds(cnxn net.Conn) {
  fmt.Printf("New connection from %v\n", cnxn.RemoteAddr())
  buf := bufio.NewReader(cnxn)

  data := make([]byte, 65535) // max size of a send we'll receive in our proto.
  var err error
  if err = magic(buf) ; err != nil {
    fmt.Fprintf(os.Stderr, "magic failure: %v\n", err)
    cnxn.Close()
    return
  }

  for {
    err := processCommand(buf, data)
    if err == io.EOF {
      fmt.Printf("client %v disconnected.\n", cnxn.RemoteAddr())
      cnxn.Close()
      return
    }
    if err != nil {
      fmt.Fprintf(os.Stderr, "could not process command: %v\n", err)
      cnxn.Close()
      return
    }
  }
}

func processCommand(buf *bufio.Reader, data []byte) (error) {
  cmd, err := buf.ReadByte()
  if err != nil {
    if err == io.EOF { return err }
    return fmt.Errorf("could not read command code from client: %v", err)
  }
  switch(cmd) {
    case cmd_OPEN:
      filename, err := readstr(buf, data)
      if err != nil {
        return err
      }
      fmt.Printf("OPEN (%d)%s\n", len(filename), filename)
      break
    case cmd_CLOSE:
      filename, err := readstr(buf, data)
      if err != nil { return err }
      fmt.Printf("CLOSE (%d)%s\n", len(filename), filename)
      break
    case cmd_BRICK:
      var lod uint32
      var bidx uint32
      err := binary.Read(buf, binary.BigEndian, &lod)
      if err != nil { return err }
      err = binary.Read(buf, binary.BigEndian, &bidx)
      if err != nil { return err }
      fmt.Printf("BRICK lod=%d, bidx=%d\n", lod, bidx)
      break
    default:
      return fmt.Errorf("unknown command: %d", uint(cmd))
  }
  return nil
}

// reads a string in our encoded way (size then string, a la fortran).
func readstr(buf *bufio.Reader, data []byte) (string, error) {
  var u16 uint16
  err := binary.Read(buf, binary.BigEndian, &u16)
  if err != nil {
    return "", fmt.Errorf("error reading string length on open: %v", err)
  }
  length := u16
  n, err := buf.Read(data[0:length])
  if err != nil {
    return "", fmt.Errorf("error reading open filename: %v", err)
  }
  if n != int(length) {
    return "", fmt.Errorf("short read for open filename: %v", err)
  }
  return string(data[0:length]), nil
}
