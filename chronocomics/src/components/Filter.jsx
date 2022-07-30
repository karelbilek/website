/* @flow */

import React from 'react';
import type {Filter as FilterData} from '../redux/state';

type StaticProps = {
    filter: FilterData;
}

type Props = StaticProps & {
    remove: () => void;
}

export class Filter extends React.PureComponent {
  props: Props;

  render() {
    return <div
      className='filter'
    >
      <span
        className='filterDesc'
      >
        {this.props.filter.type}        : {this.props.filter.query}
      </span>
      <span
        onClick={ this.props.remove }
        className='filterX'
      >        X
      </span>
    </div>;
  }
}
