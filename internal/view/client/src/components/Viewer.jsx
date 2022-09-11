import Catalog from './Catalog.jsx';
import Collection from './Collection.jsx';
import Item from './Item.jsx';
import Page from './Page.jsx';
import React from 'react';
import {getType} from '../util/stac.js';
import {useParams} from 'react-router-dom';
import {useProxy} from '../client.js';

const components = {
  Catalog: Catalog,
  Collection: Collection,
  Item: Item,
};

function Viewer() {
  const {'*': path} = useParams();
  const {data} = useProxy(path);
  if (!data) {
    return null;
  }

  const type = getType(data);
  if (!(type in components)) {
    return (
      <Page>
        <h1>Unsupported type: {type}</h1>
      </Page>
    );
  }

  const Resource = components[type];

  return (
    <Page>
      <Resource data={data} />
    </Page>
  );
}

export default Viewer;
