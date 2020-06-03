import 'package:flutter/material.dart';
import 'package:flutter_pagewise/flutter_pagewise.dart';
import 'package:catalog/generated/registry_models.pb.dart';
import 'service.dart';
import 'drawer.dart';
import 'adaptive.dart';
import 'help.dart';

class ProductListScreen extends StatelessWidget {
  final String title;
  final String projectID;
  ProductListScreen({Key key, this.title, this.projectID}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    ProductService.projectID = projectID; // HACK

    final isDesktop = isDisplayDesktop(context);

    if (isDesktop) {
      return Center(
        child: Row(
          children: [
            drawer(context),
            VerticalDivider(width: 5),
            Expanded(
              child: Scaffold(
                appBar: buildAppBar(context, isDesktop),
                body: ProductList(),
              ),
            ),
          ],
        ),
      );
    } else {
      return Scaffold(
        appBar: buildAppBar(context, isDesktop),
        body: ProductList(),
        drawer: drawer(context),
      );
    }
  }

  AppBar buildAppBar(BuildContext context, bool isDesktop) {
    return AppBar(
      automaticallyImplyLeading: !isDesktop,
      actions: <Widget>[
        ProductSearchBox(),
        IconButton(
          icon: const Icon(Icons.question_answer),
          tooltip: 'Help',
          onPressed: () {
            showHelp(context);
          },
        ),
        IconButton(
          icon: const Icon(Icons.settings),
          tooltip: 'Settings',
          onPressed: () {
            Navigator.pushNamed(context, '/settings');
          },
        ),
        IconButton(
          icon: const Icon(Icons.power_settings_new),
          tooltip: 'Log out',
          onPressed: () {
            Navigator.popUntil(context, ModalRoute.withName('/'));
          },
        ),
      ],
    );
  }
}

String routeNameForProductDetail(Product product) {
  final name = "/" + product.name.split("/").sublist(1).join("/");
  print("pushing " + name);
  return name;
}

const int pageSize = 50;
PagewiseLoadController<Product> pageLoadController;

class ProductList extends StatelessWidget {
  ProductList();

  @override
  Widget build(BuildContext context) {
    pageLoadController = PagewiseLoadController<Product>(
        pageSize: pageSize,
        pageFuture: (pageIndex) =>
            ProductService.getProductsPage(context, pageIndex));
    return Scrollbar(
      child: PagewiseListView<Product>(
        itemBuilder: this._itemBuilder,
        pageLoadController: pageLoadController,
      ),
    );
  }

  Widget _itemBuilder(context, Product entry, _) {
    return Column(
      children: <Widget>[
        GestureDetector(
          onTap: () async {
            Navigator.pushNamed(
              context,
              routeNameForProductDetail(entry),
              arguments: entry,
            );
          },
          child: ListTile(
            leading: GestureDetector(
                child: Icon(
                  Icons.bookmark_border,
                  color: Colors.black,
                ),
                onTap: () async {
                  print("save this API");
                }),
            title: Text(
              entry.displayName,
              style: TextStyle(fontSize: 18.0, fontWeight: FontWeight.bold),
            ),
            subtitle: Text(entry.description),
          ),
        ),
        Divider(thickness: 2)
      ],
    );
  }
}

class ProductSearchBox extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Container(
      width: 300,
      margin: EdgeInsets.fromLTRB(
        0,
        8,
        0,
        8,
      ),
      alignment: Alignment.centerLeft,
      color: Colors.white,
      child: TextField(
        decoration: InputDecoration(
            prefixIcon: Icon(Icons.search, color: Colors.black),
            border: InputBorder.none,
            hintText: 'Search APIs'),
        onSubmitted: (s) {
          if (s == "") {
            ProductService.filter = "";
          } else {
            ProductService.filter = "product_id.contains('$s')";
          }
          pageLoadController.reset();
        },
      ),
    );
  }
}
