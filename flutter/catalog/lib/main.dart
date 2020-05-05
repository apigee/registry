import 'package:flutter/material.dart';
import 'package:flutter_pagewise/flutter_pagewise.dart';
import 'dart:async';

import 'grpc_client.dart';
import 'package:catalog/generated/flame_models.pb.dart';
import 'package:catalog/generated/flame_service.pb.dart';
import 'package:catalog/generated/flame_service.pbgrpc.dart';

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
      home: MainScreen(title: 'API Products'),
    );
  }
}

class MainScreen extends StatelessWidget {
  final String title;
  MainScreen({Key key, this.title}) : super(key: key);

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
            onPressed: () {},
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
          children: const <Widget>[
            ListTile(
              leading: Icon(Icons.home),
              title: Text('API Hub'),
            ),
            Divider(),
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
            Divider(),
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
        ListTile(
          leading: GestureDetector(
              child: Icon(
                Icons.bookmark_border,
                color: Colors.black,
              ),
              onTap: () {
                print("save this API");
              }),
          title: Text(entry.name),
          subtitle: Text(entry.description),
        ),
        Divider()
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

class BackendService {
  static FlameClient getClient() => FlameClient(createClientChannel());

  static String filter;
  static Map<int, String> tokens;

  static Future<List<Product>> getPage(int pageIndex) {
    return BackendService._getProducts(
        parent: "projects/atlas",
        offset: pageIndex * pageSize,
        limit: pageSize);
  }

  static Future<List<Product>> _getProducts(
      {parent: String, offset: int, limit: int}) async {
    if (offset == 0) {
      tokens = Map();
    }
    print("getProducts " + (filter ?? ""));
    final client = getClient();
    final request = ListProductsRequest();
    request.parent = parent;
    request.pageSize = limit;
    if (filter != null) {
      request.filter = filter;
    }
    final token = tokens[offset];
    if (token != null) {
      request.pageToken = token;
    }
    try {
      final response = await client.listProducts(request);
      tokens[offset + limit] = response.nextPageToken;
      return response.products;
    } catch (err) {
      print('Caught error: $err');
      return null;
    }
  }
}
