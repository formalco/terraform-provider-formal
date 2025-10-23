#!/usr/bin/env node

/**
 * Simple script to insert changelog content into MDX file after frontmatter
 * Usage: node insert-changelog.js <changelog-file-path> <changelog-content-base64>
 */

const fs = require('fs');

function main() {
  const args = process.argv.slice(2);

  if (args.length !== 2) {
    console.log('Usage: node insert-changelog.js <changelog-file-path> <changelog-content-base64>');
    process.exit(1);
  }

  const [changelogPath, changelogContentBase64] = args;

  // Decode the base64 changelog content
  const changelogContent = Buffer.from(changelogContentBase64, 'base64').toString('utf8');

  // Read the existing file
  let content = fs.readFileSync(changelogPath, 'utf8');

  // Find the position after the frontmatter (after ---)
  const frontmatterRegex = /^---\n[\s\S]*?\n---\n/;
  const match = frontmatterRegex.exec(content);

  if (!match) {
    throw new Error('Could not find frontmatter in MDX file');
  }

  const insertPosition = match[0].length;

  // Insert the new changelog content after frontmatter
  content = content.slice(0, insertPosition) +
            '\n' + changelogContent + '\n\n' +
            content.slice(insertPosition);

  // Write the file back
  fs.writeFileSync(changelogPath, content);
  console.log(`✓ Updated ${changelogPath}`);
}

main();
