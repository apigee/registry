/// Dart test client for the Flame API.
import 'dart:async';
import 'dart:io';

import 'package:args/args.dart';
import 'package:grpc/grpc.dart';
import 'package:flame/generated/flame_service.pb.dart';
import 'package:flame/generated/flame_service.pbgrpc.dart';

import "package:http/http.dart" as http;
import "package:googleapis_auth/auth_io.dart";

var accountCredentials = new ServiceAccountCredentials.fromJson({
  "private_key_id": "<please fill in>",
  "private_key": "<please fill in>",
  "client_email": "<please fill in>@developer.gserviceaccount.com",
  "client_id": "<please fill in>.apps.googleusercontent.com",
  "type": "service_account"
});

var scopes = [];

Future<AccessCredentials> getCredentials() async {
  var client = new http.Client();
  return obtainAccessCredentialsViaServiceAccount(accountCredentials, scopes, client)
      .then((AccessCredentials credentials) {
    client.close();
    return credentials;
  });
}

Future<Null> main(List<String> args) async {
  if (args.length == 0) {
    print('Usage: ...');
    exit(0);
  }
  var command = args[0];
  var parser = new ArgParser();
  switch (command) {
    case "list-products":
      parser.addOption('project');
      break;
    default:
      print('Unknown command: $command');
      exit(-1);
  }

  var params;
  try {
    params = parser.parse(args.sublist(1));
  } catch (e) {
    print('$e');
    exit(-1);
  }

  final channel = new ClientChannel('localhost',
      port: 9999,
      options: const ChannelOptions(
          credentials: const ChannelCredentials.insecure()));
  final channelCompleter = Completer<void>();
  ProcessSignal.sigint.watch().listen((_) async {
    print("sigint begin");
    await channel.terminate();
    channelCompleter.complete();
    print("sigint end");
  });

  final stub = new FlameClient(channel);

  try {
    switch (command) {
      case "list-products":
        final request = ListProductsRequest();
        request.parent = params['project'];
        while (true) {
          final response = await stub.listProducts(request);
          response.products.forEach((product) => print(product.name));
          if (response.products.length == 0) {
            break;
          }
          request.pageToken = response.nextPageToken;
        }
        break;
      default:
        break;
    }
  } catch (e) {
    print('Caught error: $e');
  }
  await channel.shutdown();
  exit(0);
}
