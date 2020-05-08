import 'package:flutter/material.dart';
import 'package:catalog/generated/flame_models.pb.dart';
import 'package:catalog/product_list.dart';
import 'package:catalog/generated/flame_models.pb.dart';
import 'service.dart';

class ProductDetailWidget extends StatefulWidget {
  Product product;
  String name;

  ProductDetailWidget(this.product, this.name);
  @override
  _ProductDetailWidgetState createState() =>
      _ProductDetailWidgetState(this.product);
}

class _ProductDetailWidgetState extends State<ProductDetailWidget> {
  Product product;

  _ProductDetailWidgetState(this.product);

  @override
  Widget build(BuildContext context) {
    if (product == null) {
      // we need to fetch the product from the API
      final productFuture =
          BackendService.getProduct("projects/atlas" + widget.name);
      productFuture.then((product) {
        setState(() {
          this.product = product;
        });
        print(product);
      });
      return Scaffold(
        appBar: AppBar(
          title: Text(
            "API Details",
          ),
        ),
        body: Text("loading..."),
      );
    }

    return Scaffold(
      appBar: AppBar(
        title: Text(
          "API Details",
        ),
      ),
      body: Scrollbar(
        child: Container(
          decoration: BoxDecoration(
              //color:Colors.yellow,
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
                    style: Theme.of(context).textTheme.headline6,
                  ),
                  FlatButton(
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(4.0),
                      side: BorderSide(),
                    ),
                    color: Colors.white,
                    padding: EdgeInsets.fromLTRB(0, 0, 0, 0),
                    onPressed: () {},
                    child: Row(
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
            ],
          ),
        ),
      ),
    );
  }
}
