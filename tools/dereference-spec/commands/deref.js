const conf = new (require('conf'))()
const chalk = require('chalk')
const SwaggerParser = require("@apidevtools/swagger-parser");

function dereference (env) {
    deref(env.file);
}

async function deref(file){
    let api = await SwaggerParser.dereference(file);
    let data = JSON.stringify(api);
    console.log(data)
}

module.exports = dereference