// ABOUTME: commitlint configuration for the commit-msg hook in .pre-commit-config.yaml.
// ABOUTME: Conventional commits in pvm's house style: package types, comma-separated types, identifier-led subjects.

const types = [
  // conventional
  'build', 'chore', 'ci', 'docs', 'feat', 'fix', 'perf', 'refactor', 'revert', 'style', 'test',
  // pvm's packages and documents, as pu's history uses them
  'assess', 'conformance', 'infer', 'lexer', 'lower', 'oracle', 'parse', 'perltest', 'plan', 'psc', 'spec',
];

module.exports = {
  extends: ['@commitlint/config-conventional'],
  parserPreset: {
    // "docs, parse(scope): subject" — a comma-separated type list, as pvm writes it.
    parserOpts: { headerPattern: /^([\w-]+(?:, [\w-]+)*)(?:\(([^)]*)\))?!?: (.*)$/ },
  },
  plugins: [{
    rules: {
      'type-list-enum': ({ type }) => {
        const bad = (type || '').split(', ').filter((t) => !types.includes(t));
        return [bad.length === 0, `type must be one or more of [${types.join(', ')}], comma-separated; got: ${bad.join(', ')}`];
      },
    },
  }],
  rules: {
    'type-enum': [0],
    'type-list-enum': [2, 'always'],
    // Subjects often open with a code identifier (Session.Reparse..., ModuleCache).
    'subject-case': [0],
  },
};
