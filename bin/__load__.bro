## This file must exist for a package
@load ./main
@load ./local

redef Intel::item_expiration = 1day;

redef Intel::read_files += {
    brointelutils::otx_file
};