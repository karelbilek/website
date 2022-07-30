/* @flow */
import {applyMiddleware, createStore, compose} from 'redux';
import {persistStore, autoRehydrate} from 'redux-persist';
import {reducer} from './reducer';
import createActionBuffer from 'redux-action-buffer';
import {REHYDRATE} from 'redux-persist/constants';
import {hashEnhancer} from './hashEnhancer';
import {hashFromState, stateFromStateAndHash} from './hashFunctions';

const enhancer = compose(
  autoRehydrate(),
  applyMiddleware(
    createActionBuffer(REHYDRATE) // make sure to apply this after redux-thunk et al.
  ),
  hashEnhancer(hashFromState, stateFromStateAndHash)
);

export const store = createStore(
  reducer,
  undefined,
  enhancer
);

persistStore(store, {whitelist: ['readIssues']});

