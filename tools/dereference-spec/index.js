#! /usr/bin/env node
const { program } = require('commander')
const dereference = require('./commands/deref')

program
    .command('dereference')
    .option('-f, --file <file>', 'file path')
    .description('Dereference OpenAPI Spec')
    .action(dereference)

program.parse()