/* @flow */

import type {State, OrFilter, AndFilter, Filter} from './state';
import {getData} from '../data';

export function hashFromState(s: State): string {
  return stringFromOrFilter(s.filters);
}

export function stateFromStateAndHash(s: State, h: string): State {
  const shortened = decodeURI(h.replace(/^#/, '').replace(/\+/g, ' '));
  if (shortened === '') {
    return {
      ...s,
      filters: [[]],
    };
  }

  const filters = orFilterFromString(shortened);
  const results = getData(filters, s.sorting);
  return {
    ...s,
    filters,
    results,
  };
}

function stringFromFilter(f: Filter): string {
  const query: string = f.query;
  const replaced = query
    .replace(/&/g, '_and_')
    .replace(/\|/g, '_or_')
    .replace(/\+/g, '_plus_')
    .replace(/=/g, '_eq_');
  return (f.type + '=' + replaced).replace(/ /g, '+');
}

function stringFromAndFilter(f: AndFilter): string {
  return f.map(f => stringFromFilter(f)).join('&');
}

function stringFromOrFilter(f: OrFilter): string {
  return f.map(f => stringFromAndFilter(f)).join('|');
}

function filterFromString(s: string): Filter {
  const [type, query] = s.split('=');
  const replaced = query
    .replace(/_and_/g, '&')
    .replace(/_or_/g, '|')
    .replace(/_plus_/g, '+')
    .replace(/_eq_/g, '=');

  if (type === 'name') {
    return {type: 'name', query: replaced};
  }
  if (type === 'credit') {
    return {type: 'credit', query: replaced};
  }
  if (type === 'character') {
    return {type: 'character', query: replaced};
  }

  throw new Error('..');
}

function andFilterFromString(s: string): AndFilter {
  return s.split('&').filter(s => s !== '').map(s => filterFromString(s));
}

function orFilterFromString(s: string): OrFilter {
  return s.split('|').map(s => s === '' ? [] : andFilterFromString(s));
}
//    stateFromStateAndHash: (s: State, h: string) => State

