import 'package:catalog/product_detail.dart';
import 'package:flutter/material.dart';
import 'package:catalog/generated/flame_models.pb.dart';

import 'product_list.dart';
import 'home.dart';

void main() {
  runApp(Application());
}

class Application extends StatelessWidget {
  // This widget is the root of your application.
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Apigee Hub',
      theme: appTheme(),
      initialRoute: "/",
      onGenerateRoute: (settings) {
        // handle exact string patterns first.
        if (settings.name == "/") {
          return homePage(settings);
        }
        if (settings.name == "/settings") {
          return settingsPage(settings);
        }
        // handle regex patterns next, watch for possible ordering sensitivities
        final productDetail =
            RegExp(r"^/" + nameRegex + r"/products/" + nameRegex + r"$");
        if (productDetail.hasMatch(settings.name)) {
          return productPage(settings);
        }
        final productList = RegExp(r"^/" + nameRegex + r"/products" + r"$");
        if (productList.hasMatch(settings.name)) {
          return productsPage(settings);
        }
        final project = RegExp(r"^/" + nameRegex + r"$");
        if (project.hasMatch(settings.name)) {
          return projectPage(settings);
        }
        // if nothing matches, display a "not found" page.
        return notFoundPage(settings);
      },
    );
  }
}

const nameRegex = r"([a-zA-Z0-9-_\.]+)";

ThemeData appTheme() {
  return ThemeData(
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
  );
}

MaterialPageRoute homePage(RouteSettings settings) {
  return MaterialPageRoute(
    settings: settings,
    builder: (context) {
     return ProjectListScreen();
    },
  );
}

MaterialPageRoute settingsPage(RouteSettings settings) {
  return MaterialPageRoute(
    settings: settings,
    builder: (context) {
      return Scaffold(
        appBar: AppBar(
          title: const Text('Settings Page'),
        ),
      );
    },
  );
}

MaterialPageRoute productsPage(RouteSettings settings) {
  return MaterialPageRoute(
    settings: settings,
    builder: (context) {
      return ProductListScreen(title: 'Products');
    },
  );
}

MaterialPageRoute productPage(RouteSettings settings) {
  return MaterialPageRoute(
      settings: settings,
      builder: (context) {
        return ProductDetailWidget(settings.arguments, settings.name);
      });
}

MaterialPageRoute notFoundPage(RouteSettings settings) {
  return MaterialPageRoute(
      settings: settings,
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
}

MaterialPageRoute projectPage(RouteSettings settings) {
  return MaterialPageRoute(
      settings: settings,
      builder: (context) {
        return Scaffold(
          appBar: AppBar(
            title: const Text("project"),
          ),
          body: Center(
            child: Text("Project"),
          ),
        );
      });
}
