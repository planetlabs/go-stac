import Links from './Links.jsx';
import React from 'react';
import {Remark} from 'react-remark';
import {object} from 'prop-types';

function Catalog({data}) {
  return (
    <section>
      <h1>{data.title}</h1>
      <Remark>{data.description}</Remark>
      <Links links={data.links} />
      <footer>STAC Version {data.stac_version}</footer>
    </section>
  );
}

Catalog.propTypes = {
  data: object.isRequired,
};

export default Catalog;
