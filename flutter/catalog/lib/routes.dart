import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart' show kIsWeb;

import 'helpers/extensions.dart';

import 'pages/project_list.dart';
import 'pages/project_detail.dart';
import 'pages/product_list.dart';
import 'pages/product_detail.dart';
import 'pages/version_list.dart';
import 'pages/version_detail.dart';
import 'pages/spec_list.dart';
import 'pages/spec_detail.dart';
import 'pages/signin.dart';
import 'pages/home.dart';

const nameRegex = r"([a-zA-Z0-9-_\.]+)";

MaterialPageRoute generateRoute(RouteSettings settings) {
  if (kIsWeb) {
    if ((settings.name == "/") ||
        (currentUser == null) ||
        (currentUserIsAuthorized == false)) {
      return signInPage(settings);
    }
  } else {
    if (settings.name == "/") {
      return homePage(settings);
    }
  }
  // handle exact string patterns first.
  if (settings.name == "/projects") {
    return projectListPage(settings);
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
  final versionList = RegExp(
      r"^/" + nameRegex + r"/products/" + nameRegex + r"/versions" + r"$");
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
}

MaterialPageRoute signInPage(RouteSettings settings) {
  return MaterialPageRoute(
    settings: settings,
    builder: (context) {
      return SignInScreen();
    },
  );
}

MaterialPageRoute homePage(RouteSettings settings) {
  return MaterialPageRoute(
    settings: settings,
    builder: (context) {
      return HomeScreen();
    },
  );
}

MaterialPageRoute projectListPage(RouteSettings settings) {
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
            child: Text("You were sent to a page that doesn't exist."),
          ),
        );
      });
}
