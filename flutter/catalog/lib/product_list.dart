import 'package:flutter/material.dart';
import 'package:flutter_pagewise/flutter_pagewise.dart';
import 'package:catalog/generated/flame_models.pb.dart';
import 'service.dart';

class ProductListScreen extends StatelessWidget {
  final String title;
  final String projectID;
  ProductListScreen({Key key, this.title, this.projectID}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    ProductService.projectID = projectID; // HACK

    print("setting project ID to " + projectID);
    return Scaffold(
      appBar: AppBar(
        title: Text("API Hub: Products"),
        actions: <Widget>[
          ProductSearchBox(),
          IconButton(
            icon: const Icon(Icons.question_answer),
            tooltip: 'Help',
            onPressed: () {
              _showHelp(context);
            },
          ),
          IconButton(
            icon: const Icon(Icons.settings),
            tooltip: 'Settings',
            onPressed: () {
              Navigator.pushNamed(context, '/settings');
            },
          ),
          // TextBox(),
          IconButton(
            icon: const Icon(Icons.power_settings_new),
            tooltip: 'Log out',
            onPressed: () {},
          ),
        ],
      ),
      body: Center(
        child: ProductList(),
      ),
    );
  }

  void _showHelp(context) {
    showDialog(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: new Text("Help!"),
          content: new Text("I need some body."),
          actions: <Widget>[
            new FlatButton(
              child: new Text("Close"),
              onPressed: () {
                Navigator.of(context).pop();
              },
            ),
          ],
        );
      },
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
            subtitle: Text("Spec available, unverified."),
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
            hintText: 'Search API products'),
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
