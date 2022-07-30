/* @flow weak */

import type {State} from './state';
import { applyMiddleware } from 'redux';
import createActionBuffer from 'redux-action-buffer';

const CHANGE_HASH = '@@hashSynch/CHANGE_HASH';
export const hashEnhancer = (
    hashFromState: (s: State) => string,
    stateFromStateAndHash: (s: State, h: string) => State
) => createStore => (reducer, initialState) => {
  const store = createStore(liftReducer(reducer), initialState, applyMiddleware(createActionBuffer(CHANGE_HASH)));
  store.subscribe(() => {
    const hash = hashFromState(store.getState());
    if (window.location.hash !== hash) {
      window.location.hash = hash;
    }
  });

  window.addEventListener('hashchange', () => {
    const hash = window.location.hash;
    const savedHash = hashFromState(store.getState());
    if (savedHash !== hash) {
      store.dispatch({
        type: CHANGE_HASH,
        hash,
      });
    }
  }, false);

  store.dispatch({
    type: CHANGE_HASH,
    hash: window.location.hash,
  });

  function liftReducer(reducer) {
    return (state, action) => {
      if (action.type !== CHANGE_HASH) {
        return reducer(state, action);
      } else {
        return stateFromStateAndHash(state, action.hash);
      }
    };
  }

  return {
    ...store,
    replaceReducer: (reducer) => {
      return store.replaceReducer(liftReducer(reducer));
    },
  };
};
