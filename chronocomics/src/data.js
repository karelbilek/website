/* @flow */

import type {Filter, OrFilter} from './redux/state';

export type Issue = {
  +name: string;
  +year: number;
  +credits: $ReadOnlyArray<string>;
  +characters: $ReadOnlyArray<string>;
  +link: string;
  +id: number;
}

type Data = $ReadOnlyArray<Issue>;

const data: Data = require('./stuff.json'); // TODO replace with fetch

export function getData(orFilter: OrFilter, sorting: 'story' | 'year'): $ReadOnlyArray<Issue> {
  const resData = data.filter(issue =>
    orFilter.filter(andFilter => andFilter.length !== 0).some(andFilter =>
      andFilter.every(filter => issueSatisfiesFilter(filter, issue))
    )
  );
  if (sorting === 'story') {
    return resData;
  } else {
    return [...resData].sort((a, b) => {
      if (a.year !== b.year) {
        return a.year - b.year;
      }
      return a.id - b.id;
    });
  }
}

function issueSatisfiesFilter(f: Filter, i: Issue): boolean {
  if (f.type === 'name') {
    return i.name.toLowerCase().includes(f.query.toLowerCase());
  }
  if (f.type === 'credit') {
    return i.credits.some(credit => credit.toLowerCase() === f.query.toLowerCase());
  }
  if (f.type === 'character') {
    return i.characters.some(character => character.toLowerCase() === f.query.toLowerCase());
  }
  console.error(f);
  throw new Error('Strange filter');
}
