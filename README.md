# UDP Order Sleuth

UDP Order Sleuth is a Go-based tool designed to detect and analyze packet ordering issues in UDP multicast networks. It provides both sender and receiver functionalities to help network engineers, developers, system administrators, and operations personnel diagnose network performance and reliability.

## Features

- UDP multicast sender with configurable message interval
- UDP multicast receiver with out-of-order packet detection
- Real-time statistics reporting for both sender and receiver
- Customizable multicast address and port
- Graceful shutdown with final statistics output

## Prerequisites

To run UDP Order Sleuth, you need:

- Go 1.15 or higher
- Proper network configuration to allow UDP multicast traffic

## Installation

Clone the repository and build the project:

```bash
git clone https://github.com/yourusername/udp-order-sleuth.git
cd udp-order-sleuth
go build
```

## Usage

UDP Order Sleuth can be run in either sender or receiver mode.

### Sender Mode

To start the sender:

```bash
./udp-order-sleuth -send -addr 225.0.0.250:5001 -interval 1s
```

Options:
- `-addr`: Multicast address and port (default: 225.0.0.250:5001)
- `-interval`: Send interval (default: 1 second)

### Receiver Mode

To start the receiver:

```bash
./udp-order-sleuth -rev -addr 225.0.0.250:5001
```

Options:
- `-addr`: Multicast address and port to listen on (default: 225.0.0.250:5001)

## Example Output

Sender:
```
Sent: SequenceNumber:1:MessageContent
Sent: SequenceNumber:2:MessageContent
Sent messages: 2
...
```

Receiver:
```
Received message: SequenceNumber:6:MessageContent
OutOfOrder: Missed 5 messages. Expected 1 but received 6.
Received message: SequenceNumber:7:MessageContent
Received message: SequenceNumber:8:MessageContent
Received message: SequenceNumber:9:MessageContent
^C
Received interrupt signal. Exiting...
Stopping receiver...
Receiver stopped.
Total Received Messages: 4
Out of Order Messages: 1
```

## Limitations and Known Issues

- The tool currently only supports IPv4 multicast addresses.
- Large networks or high packet rates may affect the accuracy of out-of-order detection.
- The sender does not currently support simulating packet loss or deliberate out-of-order sending.

## Contributing

Contributions to UDP Order Sleuth are welcome! Here are some ways you can contribute:

1. Report bugs or suggest features by opening an issue.
2. Improve documentation, including this README.
3. Submit pull requests with bug fixes or new features.

Please ensure your code adheres to the existing style and includes appropriate tests.

## License

UDP Order Sleuth is released under the MIT License. See the [LICENSE](LICENSE) file for details.
