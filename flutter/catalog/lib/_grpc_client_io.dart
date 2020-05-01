import 'package:grpc/grpc.dart' as grpc;

grpc.ClientChannel createClientChannel() => grpc.ClientChannel('localhost',
    port: 9999,
    options:
        const grpc.ChannelOptions(credentials: const grpc.ChannelCredentials.insecure()));
