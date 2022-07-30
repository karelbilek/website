/* @flow */

import React from 'react';
import type {Issue as IssueData} from '../redux/state';

type StaticProps = {
    data: IssueData;
    read: boolean;
}

type Props = StaticProps & {
    toggl: () => void;
}

export class Issue extends React.PureComponent {
  props: Props;

  render() {
    return <div>
      <input
        type='checkbox'
        checked={ this.props.read }
        onChange={ this.props.toggl }
      />
      <span
        className = 'issueName'
        onClick = { this.props.toggl }
      >
        {this.props.data.name}      - {this.props.data.year}
      </span>
      &nbsp;
      <a
        href = {
              'http://www.supermegamonkey.net/chronocomic/entries/' + this.props.data.link + '.shtml'
          }
        target = '_blank'
      >
🐵
</a>
    </div>;
  }
}
