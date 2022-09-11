import Links from './Links.jsx';
import Properties from './Properties.jsx';
import React from 'react';
import {object} from 'prop-types';

function Item({data}) {
  return (
    <section>
      <h1>{data.id}</h1>
      <Properties data={data.properties} />
      <Links links={data.links} />
      <footer>STAC Version {data.stac_version}</footer>
    </section>
  );
}

Item.propTypes = {
  data: object.isRequired,
};

export default Item;
