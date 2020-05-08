import 'package:catalog/product_detail.dart';
import 'package:flutter/material.dart';
import 'package:catalog/generated/flame_models.pb.dart';

import 'product_list.dart';

void main() {
  runApp(Application());
}

class Application extends StatelessWidget {
  // This widget is the root of your application.
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
        title: 'API Products',
        theme: ThemeData(
          // Define the default brightness and colors.
          brightness: Brightness.light,
          primaryColor: Colors.orange[800],
          accentColor: Colors.pink[600],

          // Define the default font family.
          fontFamily: 'Roboto',

          // Define the default TextTheme. Use this to specify the default
          // text styling for headlines, titles, bodies of text, and more.
          textTheme: TextTheme(
            headline5: TextStyle(fontSize: 72.0, fontWeight: FontWeight.bold),
            headline6: TextStyle(fontSize: 32.0, fontWeight: FontWeight.bold),
            bodyText2: TextStyle(fontSize: 14.0),
          ),
        ),
        initialRoute: "/",
        onGenerateRoute: (settings) {
          if (settings.name == "/") {
            return MaterialPageRoute(
              settings: RouteSettings(name: "/"),
              builder: (context) {
                return ProductListScreen(title: 'API Products');
              },
            );
          }

          if (settings.name == "/settings") {
            return MaterialPageRoute(
                settings: RouteSettings(name: "/settings"),
                builder: (context) {
                  return Scaffold(
                    appBar: AppBar(
                      title: const Text('Settings Page'),
                    ),
                  );
                });
          }

          final productDetailMatch =
              RegExp(r"^/products/(\w+)").firstMatch(settings.name);
          if (productDetailMatch != null) {
            final productID = productDetailMatch.group(1);
            print("product ID: " + productID);
            print("arguments: ${settings.arguments}");

            return MaterialPageRoute(
                settings: RouteSettings(name: settings.name),
                builder: (context) {
                  final product = settings.arguments;
                  return ProductDetailWidget(product, settings.name);
                });
          }

          return MaterialPageRoute(
              settings: RouteSettings(name: settings.name),
              builder: (context) {
                return Scaffold(
                  appBar: AppBar(
                    title: const Text('NOT FOUND'),
                  ),
                  body: Center(
                    child: Text("404"),
                  ),
                );
              });
        });
  }
}
