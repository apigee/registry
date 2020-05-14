import 'package:flutter/material.dart';

import 'project_list.dart';
import 'project_detail.dart';
import 'product_list.dart';
import 'product_detail.dart';
import 'version_list.dart';
import 'version_detail.dart';
import 'spec_list.dart';
import 'spec_detail.dart';

extension ListReduction on List {
  List allButLast() {
    return this.sublist(1, this.length - 1);
  }
}

extension StringReduction on String {
  String allButLast(String separator) {
    return this.split(separator).allButLast().join(separator);
  }
}

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
        final specDetail = RegExp(r"^/" +
            nameRegex +
            r"/products/" +
            nameRegex +
            r"/versions/" +
            nameRegex +
            r"/specs/" +
            nameRegex +
            r"$");
        if (specDetail.hasMatch(settings.name)) {
          return specPage(settings);
        }
        final specList = RegExp(r"^/" +
            nameRegex +
            r"/products/" +
            nameRegex +
            r"/versions/" +
            nameRegex +
            r"/specs" +
            r"$");
        if (specList.hasMatch(settings.name)) {
          print("spec list page matched");
          return specsPage(settings);
        }
        final versionDetail = RegExp(r"^/" +
            nameRegex +
            r"/products/" +
            nameRegex +
            r"/versions/" +
            nameRegex +
            r"$");
        if (versionDetail.hasMatch(settings.name)) {
          return versionPage(settings);
        }
        final versionList = RegExp(r"^/" +
            nameRegex +
            r"/products/" +
            nameRegex +
            r"/versions" +
            r"$");
        if (versionList.hasMatch(settings.name)) {
          print("version list page matched");
          return versionsPage(settings);
        }

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
      headline1: TextStyle(fontSize: 48.0, fontWeight: FontWeight.bold),
      headline2: TextStyle(fontSize: 32.0, fontWeight: FontWeight.bold),
      headline3: TextStyle(fontSize: 24.0, fontWeight: FontWeight.bold),
      headline4: TextStyle(fontSize: 20.0, fontWeight: FontWeight.bold),
      headline5: TextStyle(fontSize: 18.0, fontWeight: FontWeight.bold),
      headline6: TextStyle(fontSize: 16.0, fontWeight: FontWeight.bold),
      bodyText1: TextStyle(fontSize: 16.0),
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

MaterialPageRoute projectPage(RouteSettings settings) {
  return MaterialPageRoute(
      settings: settings,
      builder: (context) {
        return ProjectDetailWidget(settings.arguments, settings.name);
      });
}

MaterialPageRoute productsPage(RouteSettings settings) {
  return MaterialPageRoute(
    settings: settings,
    builder: (context) {
      final projectID = settings.name.split("/")[1];
      return ProductListScreen(title: 'Products', projectID: projectID);
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

MaterialPageRoute versionsPage(RouteSettings settings) {
  return MaterialPageRoute(
    settings: settings,
    builder: (context) {
      final productID = settings.name.allButLast("/");
      print("productID = $productID");
      return VersionListScreen(title: 'Versions', productID: productID);
    },
  );
}

MaterialPageRoute versionPage(RouteSettings settings) {
  return MaterialPageRoute(
      settings: settings,
      builder: (context) {
        return VersionDetailWidget(settings.arguments, settings.name);
      });
}

MaterialPageRoute specsPage(RouteSettings settings) {
  return MaterialPageRoute(
    settings: settings,
    builder: (context) {
      final versionID = settings.name.allButLast("/");
      print("listing specs for $versionID");
      return SpecListScreen(title: 'Specs', versionID: versionID);
    },
  );
}

MaterialPageRoute specPage(RouteSettings settings) {
  return MaterialPageRoute(
      settings: settings,
      builder: (context) {
        return SpecDetailWidget(settings.arguments, settings.name);
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
