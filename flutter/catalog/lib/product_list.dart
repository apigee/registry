import 'package:flutter/material.dart';
import 'package:flutter_pagewise/flutter_pagewise.dart';
import 'service.dart';
import 'package:catalog/generated/flame_models.pb.dart';

class ProductListScreen extends StatelessWidget {
  final String title;
  ProductListScreen({Key key, this.title}) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text("API Hub"),
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
      drawer: Drawer(
        child: ListView(
          padding: EdgeInsets.zero,
          children: <Widget>[
            ListTile(
              leading: Icon(Icons.home),
              title: Text('API Hub'),
            ),
            Divider(thickness: 2),
            ListTile(
              leading: Icon(Icons.list),
              title: Text('Browse APIs'),
            ),
            ListTile(
              leading: Icon(Icons.person),
              title: Text('My APIs'),
            ),
            ListTile(
              leading: Icon(Icons.bookmark),
              title: Text('Saved APIs'),
            ),
            Center(
              child: Wrap(
                children: [
                  FlatButton(
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(20.0),
                      side: BorderSide(),
                    ),
                    color: Colors.white,
                    padding: EdgeInsets.fromLTRB(20, 0, 25, 0),
                    onPressed: () {},
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.add),
                        Text(
                          "Add API",
                          style: TextStyle(
                            fontSize: 14.0,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            Divider(thickness: 2),
            ListTile(
              leading: Icon(Icons.school),
              title: Text('API Design Process'),
            ),
            ListTile(
              leading: Icon(Icons.school),
              title: Text('API Style Guide'),
            ),
          ],
        ),
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
  final name = "/products/" + product.name.split("/").last;

  print("pushing " + name);
  return name;
}

const int pageSize = 50;
PagewiseLoadController<Product> pageLoadController;

class ProductList extends StatelessWidget {
  ProductList() {
    pageLoadController = PagewiseLoadController<Product>(
        pageSize: pageSize, pageFuture: BackendService.getPage);
  }

  @override
  Widget build(BuildContext context) {
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
            hintText: 'Search all APIs'),
        onSubmitted: (s) {
          if (s == "") {
            BackendService.filter = "";
          } else {
            BackendService.filter = "product_id.contains('$s')";
          }
          pageLoadController.reset();
        },
      ),
    );
  }
}
