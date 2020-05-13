import 'package:flutter/material.dart';
import 'package:catalog/generated/flame_models.pb.dart';
import 'service.dart';

class ProductDetailWidget extends StatefulWidget {
  final Product product;
  final String name;

  ProductDetailWidget(this.product, this.name);
  @override
  _ProductDetailWidgetState createState() =>
      _ProductDetailWidgetState(this.product);
}

class _ProductDetailWidgetState extends State<ProductDetailWidget> {
  Product product;
  List<Property> properties;

  _ProductDetailWidgetState(this.product);

  @override
  Widget build(BuildContext context) {
    final productName = "projects" + widget.name;
    if (product == null) {
      // we need to fetch the product from the API
      final productFuture = ProductService.getProduct(productName);
      productFuture.then((product) {
        setState(() {
          this.product = product;
        });
        print(product);
      });
      return Scaffold(
        appBar: AppBar(
          title: Text(
            "API Hub",
          ),
        ),
        body: Text("loading..."),
      );
    }

    var gridView = GridView.count(
        primary: false,
        padding: const EdgeInsets.all(20),
        crossAxisSpacing: 10,
        mainAxisSpacing: 10,
        crossAxisCount: 2,
        children: <Widget>[
          mycard(context),
          mycard(context),
          mycard(context),
          mycard(context),
          mycard(context)
        ]);

    return Scaffold(
      appBar: AppBar(
        title: Text(
          "API Hub",
        ),
      ),
      body: Scrollbar(
        child: Container(
          decoration: BoxDecoration(
              // color: Colors.grey,
              ),
          margin: EdgeInsets.fromLTRB(40, 20, 40, 0),
          padding: EdgeInsets.fromLTRB(0, 0, 0, 0),
          child: Column(
            children: [
              Row(
                children: [
                  Icon(Icons.bookmark_border),
                  Text(
                    product.displayName,
                    style: Theme.of(context).textTheme.headline6,
                  ),
                  Text(
                    " 1.9.1 ",
                    style: Theme.of(context).textTheme.headline4,
                  ),
                  FlatButton(
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(4.0),
                      side: BorderSide(),
                    ),
                    color: Colors.white,
                    padding: EdgeInsets.fromLTRB(0, 0, 0, 0),
                    onPressed: () {},
                    child: Column(
                      children: [
                        Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(Icons.library_books),
                            Text(
                              "Reference",
                              style: TextStyle(
                                fontSize: 12.0,
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              /*
              DefaultTabController(
                length: 3,
                child: Column(
                  children: [
                    TabBar(
                      tabs: [
                        Tab(text: "About"),
                        Tab(text: "Support"),
                        Tab(text: "Metrics"),
                        //Tab(text: "Upcoming v2.0 (Review)"),
                      ],
                    ),
                    TabBarView(children: [
                      Text("1"),
                      Text("2"),
                      Text("3"),
                    ],),
                  ],
                ),
              ),
              */

              Wrap(
                spacing: 8.0, // gap between adjacent chips
                runSpacing: 4.0, // gap between lines
                children: <Widget>[
                  Chip(
                    avatar: CircleAvatar(
                        backgroundColor: Colors.blue.shade900,
                        child: Text('AH')),
                    label: Text('Hamilton'),
                  ),
                  Chip(
                    avatar: CircleAvatar(
                        backgroundColor: Colors.blue.shade900,
                        child: Text('ML')),
                    label: Text('Lafayette'),
                  ),
                  Chip(
                    avatar: CircleAvatar(
                        backgroundColor: Colors.blue.shade900,
                        child: Text('HM')),
                    label: Text('Mulligan'),
                  ),
                  Chip(
                    avatar: CircleAvatar(
                        backgroundColor: Colors.blue.shade900,
                        child: Text('JL')),
                    label: Text('Laurens'),
                  ),
                ],
              ),
              mycard(context),
              ConstrainedBox(
                  constraints: BoxConstraints(maxWidth: 800, maxHeight: 400),
                  child: gridView),
            ],
          ),
        ),
      ),
    );
  }
}

Card mycard(BuildContext context) {
  return Card(
    child: Column(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        const ListTile(
          leading: Icon(Icons.album),
          title: Text('The Enchanted Nightingale'),
          subtitle: Text('Music by Julie Gable. Lyrics by Sidney Stein.'),
        ),
        ButtonBar(
          children: <Widget>[
            FlatButton(
              child: const Text('BUY TICKETS'),
              onPressed: () {/* ... */},
            ),
            FlatButton(
              child: const Text('LISTEN'),
              onPressed: () {/* ... */},
            ),
          ],
        ),
      ],
    ),
  );
}
