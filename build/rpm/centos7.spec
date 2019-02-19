# build with the following command:
# rpmbuild -bb
%define debug_package %{nil}

Name:       hpc-torque-helper
Version:    %{getenv:VERSION}
Release:    1%{?dist}
Summary:    A helper package for retrieving job/system information from Torque/Moab systems
License:    FIXME
URL: https://github.com/Donders-Institute/%{name}
Source0: https://github.com/Donders-Institute/%{name}/archive/%{version}.tar.gz

BuildArch: x86_64
BuildRequires: systemd

# defin the GOPATH that is created later within the extracted source code.
%define gopath %{_tmppath}/go.rpmbuild-%{name}-%{version}

%description
A helper server for retrieving torque/moab job information with leveraged privilege.

%package server-srv
Summary: the server component of the %{name} for the pbs_server node
%description server-srv
The server interfacing the torque/mom systems running on the pbs_server node to deliver job/system information to the client.

%package server-mom
Summary: the server component of the %{name} for the pbs_mom node
%description server-mom
The server interfacing the pbs_mom node to deliver job information to the client.

%package client
Summary: the client component of the %{name}
%description client
A set of client CLI tools to interact with the server for retrieving job/system information. 

%prep
%setup -q

%preun server-srv
if [ $1 -eq 0 ]; then
    echo "stopping service trqhelpd_srv ..." 
    systemctl stop trqhelpd_srv.service
    systemctl disable trqhelpd_srv.service
fi

%preun server-mom
if [ $1 -eq 0 ]; then
    echo "stopping service trqhelpd_mom ..." 
    systemctl stop trqhelpd_mom.service
    systemctl disable trqhelpd_mom.service
fi

%build
# create GOPATH structure within the source code
rm -rf %{gopath}
mkdir -p %{gopath}/src/Donders-Institute
# copy entire directory into gopath, this duplicate the source code
cp -R %{_builddir}/%{name}-%{version} %{gopath}/src/Donders-Institute/%{name}
cd %{gopath}/src/Donders-Institute/%{name}; GOPATH=%{gopath} make

%install
mkdir -p %{buildroot}/%{_sbindir}
mkdir -p %{buildroot}/%{_bindir}
mkdir -p %{buildroot}/usr/lib/systemd/system
mkdir -p %{buildroot}/etc/sysconfig
## install the service binary
install -m 755 %{gopath}/bin/trqhelpd %{buildroot}/%{_sbindir}/trqhelpd
## install files for trqhelpd_srv service
install -m 644 scripts/trqhelpd_srv.service %{buildroot}/usr/lib/systemd/system/trqhelpd_srv.service
install -m 644 scripts/trqhelpd_srv.env %{buildroot}/etc/sysconfig/trqhelpd_srv
## install files for trqhelpd_mom service
install -m 644 scripts/trqhelpd_mom.service %{buildroot}/usr/lib/systemd/system/trqhelpd_mom.service
install -m 644 scripts/trqhelpd_mom.env %{buildroot}/etc/sysconfig/trqhelpd_mom
## install files for client tools
install -m 755 %{gopath}/bin/cluster-qstat %{buildroot}/%{_bindir}/cluster-qstat
install -m 755 %{gopath}/bin/cluster-config %{buildroot}/%{_bindir}/cluster-config
install -m 755 %{gopath}/bin/cluster-tracejob %{buildroot}/%{_bindir}/cluster-tracejob
install -m 755 %{gopath}/bin/cluster-jobmeminfo %{buildroot}/%{_bindir}/cluster-jobmeminfo

%files server-srv
%{_sbindir}/trqhelpd
/usr/lib/systemd/system/trqhelpd_srv.service
/etc/sysconfig/trqhelpd_srv

%files server-mom
%{_sbindir}/trqhelpd
/usr/lib/systemd/system/trqhelpd_mom.service
/etc/sysconfig/trqhelpd_mom

%files client
%{_bindir}/cluster-qstat
%{_bindir}/cluster-config
%{_bindir}/cluster-tracejob
%{_bindir}/cluster-jobmeminfo

%post server-srv
echo "enabling service trqhelpd_srv ..."
systemctl daemon-reload
systemctl enable trqhelpd_srv.service
echo "starting service trqhelpd_srv ..."
systemctl stop trqhelpd_srv.service
systemctl start trqhelpd_srv.service

%post server-mom
echo "enabling service trqhelpd_mom ..."
systemctl daemon-reload
systemctl enable trqhelpd_mom.service
echo "starting service trqhelpd_mom ..."
systemctl stop trqhelpd_mom.service
systemctl start trqhelpd_mom.service

%postun server-srv
if [ $1 -eq 0 ]; then
    systemctl daemon-reload
fi

%postun server-mom
if [ $1 -eq 0 ]; then
    systemctl daemon-reload
fi

%clean
rm -rf %{gopath} 
rm -f %{_topdir}/SOURCES/%{version}.tar.gz
rm -rf $RPM_BUILD_ROOT

%changelog
* Fri Sep 28 2018 Hong Lee <h.lee@donders.ru.nl> - 0.7-1
- added `traceJob` command to the server
- added `cluster-tracejob` tool to the client
* Thu Sep 27 2018 Hong Lee <h.lee@donders.ru.nl> - 0.6-1
- improved RPM spec for better handling on postun for upgrade
* Thu Sep 27 2018 Hong Lee <h.lee@donders.ru.nl> - 0.5-1
- added `jobMemInfo` command to the server
- rearranged the service component for pbs_server and pbs_mom nodes
- introduced release script to make release tag on GitHub and added RPMs as extra assets
* Tue Sep 25 2018 Hong Lee <h.lee@donders.ru.nl> - 0.4-1
- added more commands to the server
- added max. connections to the server, default is 100 and changeable via the argument
- split server and client into two RPM packages
* Fri Sep 21 2018 Hong Lee <h.lee@donders.ru.nl> - 0.3-1
- added cluster-qstat, a demo for client CLI program
- improved the client-server protocol to handle multiple commands
* Thu Sep 20 2018 Hong Lee <h.lee@donders.ru.nl> - 0.2-1
- introduced environment file in /etc/sysconfig
- added more commands to the service
* Wed Sep 19 2018 Hong Lee <h.lee@donders.ru.nl> - 0.1-1
- implemented the first interface for "checkjob --xml".
