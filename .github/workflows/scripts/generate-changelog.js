#!/usr/bin/env node

/**
 * Terraform Provider Changelog Generation Script
 *
 * This script:
 * 1. Finds the latest version in the Mintlify docs changelog
 * 2. Finds ALL missing versions from git tags (v-prefixed)
 * 3. Uses git diff between tags to extract file changes
 * 4. Generates changelog via LLM using PR descriptions AND file diffs
 * 5. Processes all missing versions in a single run (oldest to newest)
 * 6. Targets mintlify-docs repo
 */

const { execSync } = require('child_process');
const fs = require('fs');
const https = require('https');
const path = require('path');

/**
 * Execute shell command and return output
 */
function exec(command) {
  return execSync(command, {
    encoding: 'utf8'
  }).trim();
}

/**
 * Make GitHub API request
 */
function githubApi(endpoint, token) {
  return new Promise((resolve, reject) => {
    const options = {
      hostname: 'api.github.com',
      path: endpoint,
      headers: {
        'Authorization': `token ${token}`,
        'User-Agent': 'Formal-TF-Provider-Changelog-Bot',
        'Accept': 'application/vnd.github.v3+json'
      }
    };

    https.get(options, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(JSON.parse(data));
        } else {
          reject(new Error(`GitHub API error: ${res.statusCode} - ${data}`));
        }
      });
    }).on('error', reject);
  });
}

/**
 * Make OpenAI API request
 */
async function openaiApi(messages, apiKey, model = 'gpt-5') {
  return new Promise((resolve, reject) => {
    const postData = JSON.stringify({
      model,
      messages
    });

    const options = {
      hostname: 'api.openai.com',
      path: '/v1/chat/completions',
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${apiKey}`,
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(postData)
      }
    };

    const req = https.request(options, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(JSON.parse(data));
        } else {
          reject(new Error(`OpenAI API error: ${res.statusCode} - ${data}`));
        }
      });
    });

    req.on('error', reject);
    req.write(postData);
    req.end();
  });
}

/**
 * Parse the latest version from a Mintlify changelog file
 */
function parseLatestVersionFromChangelog(changelogPath) {
  if (!fs.existsSync(changelogPath)) {
    console.log(`Changelog file not found: ${changelogPath}`);
    return null;
  }

  const content = fs.readFileSync(changelogPath, 'utf8');

  // Match version headers like "## 4.12.8"
  const versionRegex = /^##\s+([\d.]+)\s*$/gm;
  const matches = [...content.matchAll(versionRegex)];

  if (matches.length === 0) {
    console.log(`No versions found in changelog: ${changelogPath}`);
    return null;
  }

  // Extract all version strings
  const versions = matches.map(m => m[1]);

  // Parse versions and find the highest one
  const parseVersion = (v) => {
    const parts = v.split('.').map(Number);
    return { major: parts[0] || 0, minor: parts[1] || 0, patch: parts[2] || 0, string: v };
  };

  const compareVersions = (a, b) => {
    if (a.major !== b.major) return b.major - a.major;
    if (a.minor !== b.minor) return b.minor - a.minor;
    return b.patch - a.patch;
  };

  const parsedVersions = versions.map(parseVersion);
  parsedVersions.sort(compareVersions);

  const latestVersion = parsedVersions[0].string;
  console.log(`Found ${versions.length} versions in changelog. Latest: ${latestVersion}`);
  return latestVersion;
}

/**
 * Get all git tags matching v* pattern, sorted by version
 * Excludes test versions (tags containing '-test')
 */
function getAllVersionTags() {
  const tagsOutput = exec('git tag -l "v*" --sort=-version:refname');
  if (!tagsOutput) {
    return [];
  }

  const tags = tagsOutput.split('\n').filter(tag => tag.trim());

  // Remove 'v' prefix and filter out test versions
  return tags
    .filter(tag => !tag.includes('-test')) // Exclude test versions
    .map(tag => ({
      tag: tag,
      version: tag.substring(1) // Remove 'v' prefix
    }));
}

/**
 * Find ALL missing versions that need changelogs
 * Returns array of version pairs: [{version, previousVersion, tag, previousTag}, ...]
 * Ordered from oldest to newest (ready to process in sequence)
 */
function findAllMissingVersions(docsRepoPath) {
  const changelogPath = path.join(docsRepoPath, 'docs/changelog/terraform-provider.mdx');
  const latestInChangelog = parseLatestVersionFromChangelog(changelogPath);
  const allVersionTags = getAllVersionTags();

  if (allVersionTags.length === 0) {
    console.log('No version tags found in repository');
    return [];
  }

  const newestVersionTag = allVersionTags[0]; // Sorted newest to oldest
  const missingVersionPairs = [];

  // If no changelog exists yet, process ALL versions from oldest to newest
  if (!latestInChangelog) {
    console.log(`No changelog exists yet, will process all ${allVersionTags.length} versions`);

    // Start from oldest version (end of array)
    for (let i = allVersionTags.length - 1; i >= 0; i--) {
      const versionTag = allVersionTags[i];
      const previousVersionTag = i < allVersionTags.length - 1 ? allVersionTags[i + 1] : null;
      missingVersionPairs.push({
        version: versionTag.version,
        tag: versionTag.tag,
        previousVersion: previousVersionTag ? previousVersionTag.version : null,
        previousTag: previousVersionTag ? previousVersionTag.tag : null
      });
    }

    return missingVersionPairs;
  }

  // If changelog already has the newest version, we're done
  if (latestInChangelog === newestVersionTag.version) {
    console.log(`Changelog is already up to date at version ${newestVersionTag.version}`);
    return [];
  }

  // Find the index of the current latest version in the changelog
  const latestChangelogIndex = allVersionTags.findIndex(vt => vt.version === latestInChangelog);

  if (latestChangelogIndex === -1) {
    console.error(`Version ${latestInChangelog} from changelog not found in git tags`);
    return [];
  }

  // Build list of missing versions from oldest to newest
  for (let i = latestChangelogIndex - 1; i >= 0; i--) {
    const versionTag = allVersionTags[i];
    const previousVersionTag = allVersionTags[i + 1];
    missingVersionPairs.push({
      version: versionTag.version,
      tag: versionTag.tag,
      previousVersion: previousVersionTag.version,
      previousTag: previousVersionTag.tag
    });
  }

  // Reverse to get oldest to newest
  missingVersionPairs.reverse();

  console.log(`Found ${missingVersionPairs.length} missing versions to process`);
  if (missingVersionPairs.length > 0) {
    console.log(`Range: ${missingVersionPairs[0].version} → ${missingVersionPairs[missingVersionPairs.length - 1].version}`);
  }

  return missingVersionPairs;
}

/**
 * Get file diff between two tags
 */
function getFileDiffBetweenTags(newerTag, olderTag) {
  try {
    if (!olderTag) {
      // If no previous version, get initial files
      return exec(`git show ${newerTag} --stat --format="" | head -100`);
    }

    // Get unified diff between tags
    const diff = exec(`git diff ${olderTag}..${newerTag} --stat`);
    return diff;
  } catch (error) {
    console.error(`Error getting diff between ${olderTag} and ${newerTag}:`, error.message);
    return '';
  }
}

/**
 * Get detailed file changes (additions/deletions) for context
 * Excludes go.mod/go.sum which are very verbose
 */
function getDetailedFileDiff(newerTag, olderTag, maxLines = 300) {
  try {
    if (!olderTag) {
      return '';
    }

    // Get unified diff with limited context, excluding verbose dependency files
    const diff = exec(`git diff ${olderTag}..${newerTag} --unified=2 -- . ':(exclude)go.sum' ':(exclude)go.mod' | head -${maxLines}`);
    return diff;
  } catch (error) {
    console.error(`Error getting detailed diff:`, error.message);
    return '';
  }
}

/**
 * Find commits between two tags
 */
function findCommitsBetweenTags(newerTag, olderTag) {
  if (!olderTag) {
    // If no previous version, get all commits up to the new version
    const commits = exec(`git log ${newerTag} --oneline --no-merges | head -50`);
    return commits.split('\n').filter(line => line.trim());
  }

  // Get commits between the two versions
  const commits = exec(`git log ${olderTag}..${newerTag} --oneline --no-merges`);
  return commits.split('\n').filter(line => line.trim());
}

/**
 * Get PR details from commits
 */
async function getPRDetailsFromCommits(commits, repo, token) {
  console.log(`Finding PRs for ${commits.length} commits...`);

  const prNumbers = new Set();

  // Extract PR numbers from commit messages
  commits.forEach(commit => {
    const matches = commit.match(/#(\d+)/g);
    if (matches) {
      matches.forEach(match => {
        prNumbers.add(match.substring(1));
      });
    }
  });

  console.log(`Found ${prNumbers.size} unique PR numbers`);

  // Fetch PR details
  const prDetails = [];
  for (const prNumber of prNumbers) {
    try {
      const pr = await githubApi(`/repos/${repo}/pulls/${prNumber}`, token);

      prDetails.push({
        number: pr.number,
        title: pr.title,
        body: pr.body || '',
        labels: pr.labels.map(l => l.name.toLowerCase())
      });
    } catch (error) {
      console.error(`Error fetching PR #${prNumber}:`, error.message);
    }
  }

  console.log(`Found ${prDetails.length} PRs`);
  return prDetails;
}

/**
 * Generate changelog content using OpenAI with file diffs
 */
async function generateChangelogWithLLM(version, prDetails, fileDiff, detailedDiff, openaiKey, date) {
  const systemPrompt = `You are a technical writer creating an external-facing changelog for our Terraform Provider customers. Your goal is to communicate **high-level functionality changes**, not internal implementation details.

### Changelog Format:
- Each release must include only the relevant sections:
  - **### New** (for new features)
  - **### Fixed** (for bug fixes)
  - **### Changed** (for modifications to existing functionality)
- Omit any section that has no changes.
- Each bullet point should be **short, clear, and impact-focused**.
- Keep bullet points between 15-120 characters.
- Use a single line per change.

### What to Include:
1. Extract customer-facing changes from PR descriptions AND file diffs
2. **Choose only one category per change** (do not duplicate across sections)
3. **Explain why the change matters** to users, avoiding technical jargon
4. **Do not include PR numbers, internal function names, or implementation details**
5. **Start each bullet point with a verb** in present tense (e.g., 'Add', 'Fix', 'Update')
6. **Group related changes** under a single bullet point to avoid fragmentation

### Example Output:
### New
- Add support for new resource type formal_policy_template

### Fixed
- Fix issue with datastore connection timeout configuration

### Changed
- Improve resource creation validation for better error messages

Focus on how changes affect Terraform Provider users, **not on how they were implemented.**`;

  // Truncate PR bodies to avoid token limits
  const truncatedPRDetails = prDetails.map(pr => ({
    ...pr,
    body: pr.body ? pr.body.substring(0, 500) : '' // Limit PR body to 500 chars
  }));

  const userPrompt = `Generate a changelog for Terraform Provider version ${version}.

Pull Request Information:
${JSON.stringify(truncatedPRDetails, null, 2)}

File Changes Summary:
${fileDiff.substring(0, 1500)}

${detailedDiff ? `\nDetailed Changes (sample):\n${detailedDiff.substring(0, 1500)}` : ''}`;

  try {
    const response = await openaiApi([
      { role: 'system', content: systemPrompt },
      { role: 'user', content: userPrompt }
    ], openaiKey);

    let changelogContent = response.choices[0].message.content;

    // Remove title if it starts with # but not with ###
    if (changelogContent.startsWith('#') && !changelogContent.startsWith('###')) {
      const firstSectionIndex = changelogContent.indexOf('### ');
      if (firstSectionIndex !== -1) {
        changelogContent = changelogContent.substring(firstSectionIndex);
      }
    }

    // Determine tags based on changelog content
    const tags = [];
    if (changelogContent.includes('### New')) tags.push('New Features');
    if (changelogContent.includes('### Fixed')) tags.push('Bug Fixes');
    if (changelogContent.includes('### Changed')) tags.push('Improvements');

    const tagsString = tags.length > 0 ? ` tags={${JSON.stringify(tags)}}` : '';

    // Format date as "Month Day, Year"
    const formatDate = (dateStr) => {
      const d = new Date(dateStr);
      const monthNames = ['January', 'February', 'March', 'April', 'May', 'June',
                         'July', 'August', 'September', 'October', 'November', 'December'];
      return `${monthNames[d.getMonth()]} ${d.getDate()}, ${d.getFullYear()}`;
    };

    const formattedDate = formatDate(date);

    // Return with Update wrapper for Mintlify
    return `<Update label="${formattedDate}"${tagsString}>

## ${version}

${changelogContent.trim()}

</Update>`;

  } catch (error) {
    console.error('Error generating changelog with LLM:', error.message);
    throw error;
  }
}

/**
 * Insert changelog into docs file (Mintlify MDX format)
 */
function insertChangelogIntoFile(changelogPath, changelogContent) {
  // Create file if it doesn't exist
  if (!fs.existsSync(changelogPath)) {
    const dir = path.dirname(changelogPath);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }

    // Create Mintlify-style changelog file
    const header = `---
title: "Terraform Provider Changelog"
description: "Updates and releases for the Formal Terraform Provider"
---

`;
    fs.writeFileSync(changelogPath, header);
  }

  let content = fs.readFileSync(changelogPath, 'utf8');

  // Find the position after the frontmatter (after ---)
  const frontmatterRegex = /^---\n[\s\S]*?\n---\n/;
  const match = frontmatterRegex.exec(content);

  if (!match) {
    throw new Error('Could not find frontmatter in MDX file');
  }

  const insertPosition = match.index + match[0].length;

  // Insert the new changelog content after frontmatter
  content = content.slice(0, insertPosition) +
            '\n' + changelogContent + '\n\n' +
            content.slice(insertPosition);

  fs.writeFileSync(changelogPath, content);
  console.log(`✓ Updated ${changelogPath}`);
}

/**
 * Main function
 */
async function main() {
  const args = process.argv.slice(2);

  // Parse optional arguments
  let docsRepoPath = 'docs-repo';
  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--docs-repo-path' && i + 1 < args.length) {
      docsRepoPath = args[i + 1];
      i++;
    }
  }

  // Required environment variables
  const githubToken = process.env.GITHUB_TOKEN;
  const githubRepo = process.env.GITHUB_REPOSITORY || 'formalco/terraform-provider-formal';
  const openaiKey = process.env.OPENAI_API_KEY;

  if (!githubToken) {
    console.error('GITHUB_TOKEN environment variable is required');
    process.exit(1);
  }

  if (!openaiKey) {
    console.error('OPENAI_API_KEY environment variable is required');
    process.exit(1);
  }

  console.log('='.repeat(70));
  console.log('Generating changelog for Terraform Provider');
  console.log('='.repeat(70));

  // Step 1: Find ALL missing versions needing changelogs
  console.log('\n[1/6] Finding all missing versions...');
  const missingVersionPairs = findAllMissingVersions(docsRepoPath);

  if (missingVersionPairs.length === 0) {
    console.log('No versions need changelogs. Exiting.');
    process.exit(0);
  }

  console.log(`\nWill process ${missingVersionPairs.length} versions in order (oldest to newest):\n`);
  missingVersionPairs.forEach((v, i) => {
    console.log(`  ${i + 1}. ${v.version}${v.previousVersion ? ` (vs ${v.previousVersion})` : ' (first version)'}`);
  });
  console.log();

  const changelogPath = path.join(docsRepoPath, 'docs/changelog/terraform-provider.mdx');
  const allGeneratedChangelogs = [];
  const allPRsByVersion = [];

  // Process each version in sequence
  for (let i = 0; i < missingVersionPairs.length; i++) {
    const { version, tag, previousVersion, previousTag } = missingVersionPairs[i];

    console.log('\n' + '━'.repeat(70));
    console.log(`Processing version ${i + 1}/${missingVersionPairs.length}: ${version}`);
    console.log('━'.repeat(70));

    // Get the tag date for the version in UTC
    const tagDate = exec(`TZ=UTC git log -1 --format=%cd --date=format-local:%Y-%m-%d ${tag}`);
    console.log(`Version ${version} date (UTC): ${tagDate}`);

    // Step 2: Get file diff between versions
    console.log('\n[2/6] Getting file diff between versions...');
    const fileDiff = getFileDiffBetweenTags(tag, previousTag);
    const detailedDiff = getDetailedFileDiff(tag, previousTag);
    console.log(`File diff stats:\n${fileDiff.substring(0, 500)}`);

    // Step 3: Find commits between versions
    console.log('\n[3/6] Finding commits between versions...');
    const commits = findCommitsBetweenTags(tag, previousTag);
    console.log(`Found ${commits.length} commits`);

    // Step 4: Get PR details
    console.log('\n[4/6] Fetching PR details...');
    const prDetails = await getPRDetailsFromCommits(commits, githubRepo, githubToken);

    // Store PR information for this version
    if (prDetails.length > 0) {
      allPRsByVersion.push({
        version,
        prs: prDetails.map(pr => ({ number: pr.number, title: pr.title }))
      });
    }

    // Step 5: Generate changelog
    console.log('\n[5/6] Generating changelog with LLM...');
    const changelogContent = await generateChangelogWithLLM(
      version,
      prDetails,
      fileDiff,
      detailedDiff,
      openaiKey,
      tagDate
    );

    console.log('\n' + '-'.repeat(70));
    console.log(`Generated Changelog for ${version}:`);
    console.log('-'.repeat(70));
    console.log(changelogContent);
    console.log('-'.repeat(70));

    allGeneratedChangelogs.push(changelogContent);

    console.log(`✅ Version ${version} complete!\n`);
  }

  // Step 6: Insert all changelogs at once
  console.log('\n[6/6] Inserting changelogs into file...');
  console.log(`Inserting ${allGeneratedChangelogs.length} changelogs...`);

  // Reverse the array so newest versions appear first in the file
  // (each insert adds after frontmatter, so we insert newest first)
  allGeneratedChangelogs.reverse();

  for (let i = 0; i < allGeneratedChangelogs.length; i++) {
    insertChangelogIntoFile(changelogPath, allGeneratedChangelogs[i]);
  }

  console.log('\n' + '='.repeat(70));
  console.log(`✅ Successfully generated ${allGeneratedChangelogs.length} changelogs!`);
  console.log('='.repeat(70));

  // Output for GitHub Actions
  if (process.env.GITHUB_OUTPUT) {
    const newestVersion = missingVersionPairs[0].version;
    const oldestVersion = missingVersionPairs[missingVersionPairs.length - 1].version;
    const versionRange = missingVersionPairs.length === 1 ? newestVersion : `${oldestVersion}-to-${newestVersion}`;

    fs.appendFileSync(process.env.GITHUB_OUTPUT, `version=${versionRange}\n`);
    fs.appendFileSync(process.env.GITHUB_OUTPUT, `versions_count=${allGeneratedChangelogs.length}\n`);

    // Create PR list for the GitHub PR description (newest first)
    let prListMarkdown = '';
    const reversedPRsByVersion = [...allPRsByVersion].reverse();
    for (const { version, prs } of reversedPRsByVersion) {
      prListMarkdown += `\n### ${version}\n`;
      for (const pr of prs) {
        prListMarkdown += `- [#${pr.number}](https://github.com/${githubRepo}/pull/${pr.number}): ${pr.title}\n`;
      }
    }

    // Write PR list to a temporary file
    const prListPath = '/tmp/pr-list.md';
    fs.writeFileSync(prListPath, prListMarkdown);
    fs.appendFileSync(process.env.GITHUB_OUTPUT, `pr_list_file=${prListPath}\n`);
  }
}

// Run main function
if (require.main === module) {
  main().catch(error => {
    console.error('Error:', error);
    process.exit(1);
  });
}

module.exports = { main };
