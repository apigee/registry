import 'package:flutter/material.dart';
import 'package:catalog/generated/registry_models.pb.dart';
import 'service.dart';

class ProductDetailWidget extends StatefulWidget {
  final Product product;
  final String name;

  ProductDetailWidget(this.product, this.name);
  @override
  _ProductDetailWidgetState createState() =>
      _ProductDetailWidgetState(this.product);
}

String routeNameForProductDetailVersions(Product product) {
  final name = "/" + product.name.split("/").sublist(1).join("/") + "/versions";
  print("pushing " + name);
  return name;
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

    return Scaffold(
      appBar: AppBar(
        title: Text(
          "API Hub",
        ),
      ),
      body: SingleChildScrollView(
        child: Center(
          child: Column(
            children: [
              Row(children: [productCard(context, product)]),
              Row(children: [
                productInfoCard(context, product),
                productInfoCard(context, product)
              ]),
              Row(children: [
                productInfoCard(context, product),
                productInfoCard(context, product),
                productInfoCard(context, product)
              ]),
            ],
          ),
        ),
      ),
    );
  }
}

Expanded productCard(BuildContext context, Product product) {
  return Expanded(
    child: Card(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          ListTile(
            leading: Icon(Icons.album),
            title: Text(product.displayName,
                style: Theme.of(context).textTheme.headline5),
            subtitle: Text(product.description),
          ),
          ButtonBar(
            children: <Widget>[
              FlatButton(
                child: const Text('VERSIONS'),
                onPressed: () {
                  Navigator.pushNamed(
                    context,
                    routeNameForProductDetailVersions(product),
                    arguments: product,
                  );
                },
              ),
              FlatButton(
                child: const Text('MORE'),
                onPressed: () {/* ... */},
              ),
            ],
          ),
        ],
      ),
    ),
  );
}

Expanded productInfoCard(BuildContext context, Product product) {
  return Expanded(
    child: Card(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          ListTile(
            leading: Icon(Icons.album),
            title: Text("More Info",
                style: Theme.of(context).textTheme.headline6),
            subtitle: Text("$product"),
          ),
          ButtonBar(
            children: <Widget>[             
              FlatButton(
                child: const Text('MORE'),
                onPressed: () {/* ... */},
              ),
            ],
          ),
        ],
      ),
    ),
  );
}
