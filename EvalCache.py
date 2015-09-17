#!python

import os , sys , json
from argparse import ArgumentParser


#Accept path arguement to list folder contents of a directory and path to create json file

def add_argparse_group(parser):
    """
    Given an ArgumentParser, tell it the expected arguments
    (and corresponding defaults, helpstrings, etc).
    """
    parser.add_argument('-c','--cachepath', type=str, help='The full path to the Cache location.', dest='cachepath', default='/tmp' )
    parser.add_argument('-j','--jsonpath', type=str, help='The full path for the Json file to be created.', dest='jsonpath', default='/tmp/volcache.json' )



def get_folders():
    parser = ArgumentParser('get_folders')
    add_argparse_group(parser)
    args = parser.parse_args()
    subdirectories = sorted(os.listdir(args.cachepath))
    return build_schema(subdirectories)

#Append list to dictionary and provide host info, tag volumes - data0, data1, etc in an alphabetically ordered dict

def build_schema(subdirectories):
    parser = ArgumentParser('get_folders')
    add_argparse_group(parser)
    args = parser.parse_args()
    schema=[]
    for el in subdirectories:
        info = dict(Host=el, Path=args.cachepath, ID=subdirectories.index(el))
        schema.append(info)
    return schema_export(schema)

#print build_schema.info

def schema_export(schema):
    parser = ArgumentParser('get_folders')
    add_argparse_group(parser)
    args = parser.parse_args()
    dest=args.jsonpath
#    jfile=json.dumps(schema)
    with open(dest, 'w') as outfile:
            json.dump(schema, outfile)
    return True

# {'index': 1, 'id': '2345', 'name': 'Tom'}

# Use defined json file name/path and output to file.


#json.dumps(subdirectories)

if __name__ == '__main__':
    get_folders()