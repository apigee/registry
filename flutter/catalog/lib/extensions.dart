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